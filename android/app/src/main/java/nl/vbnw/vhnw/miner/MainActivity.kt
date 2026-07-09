package nl.vbnw.vhnw.miner

import android.content.Context
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.os.PowerManager
import android.provider.Settings
import android.widget.Button
import android.widget.CheckBox
import android.widget.EditText
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity

class MainActivity : AppCompatActivity() {

    private val handler = Handler(Looper.getMainLooper())
    private lateinit var statusText: TextView
    private lateinit var startStopButton: Button

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        statusText = findViewById(R.id.statusText)
        startStopButton = findViewById(R.id.startStopButton)

        loadIntoFields(Config.load(this))

        startStopButton.setOnClickListener {
            if (MiningService.isRunning) {
                MiningService.stop(this)
            } else {
                Config.save(this, readFromFields())
                MiningService.start(this)
            }
            refreshStatus()
        }

        findViewById<Button>(R.id.saveButton).setOnClickListener {
            Config.save(this, readFromFields())
            Toast.makeText(this, "Instellingen opgeslagen", Toast.LENGTH_SHORT).show()
        }

        findViewById<Button>(R.id.batteryButton).setOnClickListener {
            requestIgnoreBatteryOptimizations()
        }
    }

    override fun onResume() {
        super.onResume()
        refreshStatus()
    }

    /** Simpele polling i.p.v. een bound service — voldoende voor dit statusscherm. */
    private fun refreshStatus() {
        val running = MiningService.isRunning
        startStopButton.text = if (running) "STOP" else "START"
        val summary = MiningService.lastSummary
        statusText.text = when {
            !running -> "Gestopt"
            summary != null -> "%s · %.0f H/s · %d shares · %d herstarts"
                .format(MiningService.lastState, summary.hashrateNow, summary.sharesGood, MiningService.restarts.get())
            else -> MiningService.lastState
        }
        handler.postDelayed({ if (!isFinishing) refreshStatus() }, 3000)
    }

    private fun requestIgnoreBatteryOptimizations() {
        val pm = getSystemService(Context.POWER_SERVICE) as PowerManager
        if (pm.isIgnoringBatteryOptimizations(packageName)) {
            Toast.makeText(this, "Al uitgezet — deze app mag op de achtergrond blijven draaien", Toast.LENGTH_SHORT).show()
            return
        }
        val intent = android.content.Intent(Settings.ACTION_REQUEST_IGNORE_BATTERY_OPTIMIZATIONS)
        intent.data = Uri.parse("package:$packageName")
        startActivity(intent)
    }

    private fun loadIntoFields(c: Config) {
        findViewById<EditText>(R.id.poolInput).setText(c.pool)
        findViewById<EditText>(R.id.walletInput).setText(c.wallet)
        findViewById<EditText>(R.id.workerInput).setText(c.worker)
        findViewById<EditText>(R.id.referralInput).setText(c.referral)
        findViewById<EditText>(R.id.threadsInput).setText(c.threads.toString())
        findViewById<CheckBox>(R.id.tlsCheck).isChecked = c.tls
        findViewById<CheckBox>(R.id.autoStartCheck).isChecked = c.autoStart
        findViewById<CheckBox>(R.id.communityCheck).isChecked = c.communityEnabled
        findViewById<EditText>(R.id.hubUrlInput).setText(c.hubUrl)
        findViewById<EditText>(R.id.deviceNameInput).setText(c.deviceName)
        findViewById<EditText>(R.id.ownerInput).setText(c.owner)
        findViewById<EditText>(R.id.tokenInput).setText(c.token)
    }

    private fun readFromFields(): Config {
        val current = Config.load(this)
        return current.copy(
            pool = findViewById<EditText>(R.id.poolInput).text.toString(),
            wallet = findViewById<EditText>(R.id.walletInput).text.toString(),
            worker = findViewById<EditText>(R.id.workerInput).text.toString(),
            referral = findViewById<EditText>(R.id.referralInput).text.toString(),
            threads = findViewById<EditText>(R.id.threadsInput).text.toString().toIntOrNull() ?: 0,
            tls = findViewById<CheckBox>(R.id.tlsCheck).isChecked,
            autoStart = findViewById<CheckBox>(R.id.autoStartCheck).isChecked,
            communityEnabled = findViewById<CheckBox>(R.id.communityCheck).isChecked,
            hubUrl = findViewById<EditText>(R.id.hubUrlInput).text.toString(),
            deviceName = findViewById<EditText>(R.id.deviceNameInput).text.toString().ifBlank { Build.MODEL ?: "android" },
            owner = findViewById<EditText>(R.id.ownerInput).text.toString(),
            token = findViewById<EditText>(R.id.tokenInput).text.toString(),
        )
    }
}
