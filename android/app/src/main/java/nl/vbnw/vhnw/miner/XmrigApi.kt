package nl.vbnw.vhnw.miner

import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL

/** Statistieken uit xmrig's ingebouwde HTTP API (/2/summary), net als bij de Go-agent. */
data class XmrigSummary(
    val version: String,
    val hashrateNow: Double,
    val sharesGood: Long,
    val sharesTotal: Long,
    val uptimeSecs: Long,
    val pool: String,
)

class XmrigApi(private val apiPort: Int) {

    /** Haalt /2/summary op van de lokaal draaiende xmrig. Geeft null als hij (nog) niet reageert. */
    fun summary(): XmrigSummary? {
        return try {
            val url = URL("http://127.0.0.1:$apiPort/2/summary")
            val conn = url.openConnection() as HttpURLConnection
            conn.connectTimeout = 2000
            conn.readTimeout = 2000
            conn.requestMethod = "GET"
            if (conn.responseCode != 200) return null
            val body = conn.inputStream.bufferedReader().readText()
            parse(body)
        } catch (e: Exception) {
            null
        }
    }

    private fun parse(body: String): XmrigSummary {
        val root = JSONObject(body)
        val hashrateTotal = root.optJSONObject("hashrate")?.optJSONArray("total")
        val hashrateNow = if (hashrateTotal != null && !hashrateTotal.isNull(0)) {
            hashrateTotal.optDouble(0, 0.0)
        } else 0.0
        val results = root.optJSONObject("results")
        val connection = root.optJSONObject("connection")
        return XmrigSummary(
            version = root.optString("version", ""),
            hashrateNow = hashrateNow,
            sharesGood = results?.optLong("shares_good", 0) ?: 0,
            sharesTotal = results?.optLong("shares_total", 0) ?: 0,
            uptimeSecs = root.optLong("uptime", 0),
            pool = connection?.optString("pool", "") ?: "",
        )
    }
}
