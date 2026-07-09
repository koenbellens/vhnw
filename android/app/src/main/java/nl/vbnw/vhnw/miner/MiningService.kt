package nl.vbnw.vhnw.miner

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.os.BatteryManager
import android.os.Build
import android.os.IBinder
import android.os.PowerManager
import android.os.SystemClock
import androidx.core.app.NotificationCompat
import java.io.BufferedReader
import java.io.InputStreamReader
import java.util.concurrent.atomic.AtomicBoolean
import java.util.concurrent.atomic.AtomicInteger

/**
 * Foreground service die xmrig beheert, net als de watchdog in de Go-agent
 * (internal/miner/manager.go): start het proces, herstart automatisch bij een
 * crash, en meldt periodiek de status aan hash.vbnw.nl. Draait als
 * foreground-service met een permanente notificatie zodat Android hem niet
 * afschiet (dit was precies het probleem met de oude losse XMRig-apps).
 */
class MiningService : Service() {

    companion object {
        const val ACTION_START = "nl.vbnw.vhnw.miner.action.START"
        const val ACTION_STOP = "nl.vbnw.vhnw.miner.action.STOP"
        private const val CHANNEL_ID = "vhnw_mining"
        private const val NOTIF_ID = 1
        private const val API_PORT = 18088

        // Door de UI uit te lezen (eenvoudige polling i.p.v. bound-service-boilerplate).
        @Volatile var isRunning = false
        @Volatile var lastState = "gestopt"
        @Volatile var lastSummary: XmrigSummary? = null
        @Volatile var lastBatteryTempC = 0.0
        val restarts = AtomicInteger(0)

        fun start(context: Context) {
            val i = Intent(context, MiningService::class.java).setAction(ACTION_START)
            context.startForegroundService(i)
        }

        fun stop(context: Context) {
            context.startService(Intent(context, MiningService::class.java).setAction(ACTION_STOP))
        }
    }

    private val running = AtomicBoolean(false)
    private var wakeLock: PowerManager.WakeLock? = null
    private var supervisorThread: Thread? = null
    private var reporterThread: Thread? = null
    @Volatile private var process: Process? = null
    private var startedAtMs: Long = 0L

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onCreate() {
        super.onCreate()
        createNotificationChannel()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_STOP -> {
                stopMining()
                return START_NOT_STICKY
            }
            else -> startMining()
        }
        return START_STICKY
    }

    override fun onDestroy() {
        stopMining()
        super.onDestroy()
    }

    private fun startMining() {
        if (running.get()) return
        val config = Config.load(this)
        if (config.wallet.isBlank()) {
            lastState = "fout: wallet ontbreekt"
            startForeground(NOTIF_ID, buildNotification("Instellingen onvolledig: vul je wallet in"))
            return
        }

        running.set(true)
        isRunning = true
        startedAtMs = SystemClock.elapsedRealtime()
        restarts.set(0)
        lastState = "gestart"

        wakeLock = (getSystemService(Context.POWER_SERVICE) as PowerManager)
            .newWakeLock(PowerManager.PARTIAL_WAKE_LOCK, "vhnw:mining")
            .apply { setReferenceCounted(false); acquire() }

        startForeground(NOTIF_ID, buildNotification("Wordt gestart…"))

        supervisorThread = Thread { supervisorLoop(config) }.apply { isDaemon = true; start() }
        reporterThread = Thread { reporterLoop(config) }.apply { isDaemon = true; start() }
    }

    private fun stopMining() {
        running.set(false)
        isRunning = false
        lastState = "gestopt"
        process?.destroy()
        process = null
        wakeLock?.let { if (it.isHeld) it.release() }
        wakeLock = null
        stopForeground(STOP_FOREGROUND_REMOVE)
        stopSelf()
    }

    /** Start xmrig, wacht tot hij stopt, en herstart na een korte pauze — zolang de service actief is. */
    private fun supervisorLoop(config: Config) {
        val binary = applicationInfo.nativeLibraryDir + "/libxmrig.so"
        while (running.get()) {
            try {
                val args = buildArgs(config)
                val pb = ProcessBuilder(listOf(binary) + args)
                pb.redirectErrorStream(true)
                val p = pb.start()
                process = p

                // Stdout leegtrekken zodat het proces niet blokkeert op een volle pipe-buffer.
                val reader = BufferedReader(InputStreamReader(p.inputStream))
                Thread { try { while (reader.readLine() != null) { /* niets bewaren, alleen leegtrekken */ } } catch (_: Exception) {} }
                    .apply { isDaemon = true; start() }

                p.waitFor()
            } catch (e: Exception) {
                lastState = "fout: " + (e.message ?: "kon xmrig niet starten")
            }
            process = null
            if (running.get()) {
                restarts.incrementAndGet()
                lastState = "gestart"
                Thread.sleep(5000)
            }
        }
    }

    /** Elke paar seconden: stats ophalen, notificatie bijwerken, heartbeat naar de hub sturen. */
    private fun reporterLoop(config: Config) {
        val api = XmrigApi(API_PORT)
        val hub = HubReporter(config)
        val intervalMs = 15_000L
        // Geef xmrig even de tijd om op te starten voor de eerste poll.
        Thread.sleep(3000)
        while (running.get()) {
            val summary = api.summary()
            lastSummary = summary
            lastBatteryTempC = readBatteryTempC()
            val uptimeSecs = (SystemClock.elapsedRealtime() - startedAtMs) / 1000

            updateNotification(summary)
            hub.send(state = if (summary != null) "actief" else lastState, summary = summary, uptimeSecs = uptimeSecs, restarts = restarts.get())

            Thread.sleep(intervalMs)
        }
    }

    private fun buildArgs(config: Config): List<String> {
        val user = buildString {
            append(config.wallet)
            if (config.worker.isNotBlank()) append(".").append(config.worker)
            if (config.referral.isNotBlank()) append("#").append(config.referral)
        }
        val args = mutableListOf(
            "-a", config.algo,
            "-o", config.pool,
            "-u", user,
            "-p", config.password,
            "-k",
            "--http-host", "127.0.0.1",
            "--http-port", API_PORT.toString(),
            "--no-color",
        )
        if (config.tls) args += "--tls"
        if (config.threads > 0) args += listOf("-t", config.threads.toString())
        return args
    }

    private fun readBatteryTempC(): Double {
        return try {
            val intent = registerReceiver(null, IntentFilter(Intent.ACTION_BATTERY_CHANGED))
            val tenthsOfDegree = intent?.getIntExtra(BatteryManager.EXTRA_TEMPERATURE, -1) ?: -1
            if (tenthsOfDegree >= 0) tenthsOfDegree / 10.0 else 0.0
        } catch (e: Exception) {
            0.0
        }
    }

    private fun createNotificationChannel() {
        val channel = NotificationChannel(
            CHANNEL_ID, "VHNW mining", NotificationManager.IMPORTANCE_LOW
        ).apply { description = "Toont of deze telefoon aan het minen is" }
        (getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager)
            .createNotificationChannel(channel)
    }

    private fun buildNotification(text: String): Notification {
        val openApp = PendingIntent.getActivity(
            this, 0, Intent(this, MainActivity::class.java),
            PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )
        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle("VHNW miner")
            .setContentText(text)
            .setSmallIcon(R.drawable.ic_notification)
            .setOngoing(true)
            .setContentIntent(openApp)
            .build()
    }

    private fun updateNotification(summary: XmrigSummary?) {
        val text = if (summary != null) {
            "%.0f H/s · %d shares".format(summary.hashrateNow, summary.sharesGood)
        } else {
            "Bezig met opstarten…"
        }
        (getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager)
            .notify(NOTIF_ID, buildNotification(text))
    }
}
