package nl.vbnw.vhnw.miner

import android.os.Build
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL

/**
 * Stuurt een heartbeat naar de VHNW-hub (hash.vbnw.nl), in exact hetzelfde
 * JSON-formaat als de Go-agent (zie internal/community/community.go), zodat
 * telefoons gewoon meedoen in hetzelfde dashboard als de Pi/laptop.
 */
class HubReporter(private val config: Config) {

    fun send(state: String, summary: XmrigSummary?, uptimeSecs: Long, restarts: Int) {
        if (!config.communityEnabled || config.hubUrl.isBlank()) return
        try {
            val body = JSONObject().apply {
                put("device_id", config.deviceId)
                put("name", config.deviceName)
                put("owner", config.owner)
                put("type", "cpu-agent") // telefoons delen voorlopig het CPU-badge
                put("miner", if (summary != null) "XMRig ${summary.version}" else "")
                put("algo", config.algo)
                put("pool", summary?.pool?.ifBlank { config.pool } ?: config.pool)
                put("state", state)
                put("hashrate", summary?.hashrateNow ?: 0.0)
                put("shares_good", summary?.sharesGood ?: 0)
                put("shares_total", summary?.sharesTotal ?: 0)
                put("temp_c", batteryTempC())
                put("uptime_secs", uptimeSecs)
                put("restarts", restarts)
                put("agent_version", "android-0.1.0 (${Build.MODEL})")
            }

            val url = URL(config.hubUrl.trimEnd('/') + "/api/report")
            val conn = url.openConnection() as HttpURLConnection
            conn.requestMethod = "POST"
            conn.doOutput = true
            conn.connectTimeout = 5000
            conn.readTimeout = 5000
            conn.setRequestProperty("Content-Type", "application/json")
            if (config.token.isNotBlank()) {
                conn.setRequestProperty("Authorization", "Bearer ${config.token}")
            }
            conn.outputStream.use { it.write(body.toString().toByteArray()) }
            conn.responseCode // voert het verzoek uit; we negeren het resultaat verder
            conn.disconnect()
        } catch (e: Exception) {
            // Stil falen: één gemiste heartbeat is geen probleem, de volgende volgt over interval_seconds.
        }
    }

    /** Batterijtemperatuur als benadering van "temp_c" — telefoons hebben geen CPU-sensor-API. */
    private fun batteryTempC(): Double = MiningService.lastBatteryTempC
}
