package nl.vbnw.vhnw.miner

import android.content.Context
import android.content.SharedPreferences
import java.util.UUID

/**
 * Instellingen voor deze telefoon/tablet. Veldnamen volgen bewust dezelfde
 * indeling als config.example.json van de Go-agent (Pi/laptop), zodat het
 * dashboard op hash.vbnw.nl geen onderscheid hoeft te maken.
 */
data class Config(
    val pool: String,
    val algo: String,
    val tls: Boolean,
    val wallet: String,
    val worker: String,
    val referral: String,
    val password: String,
    val threads: Int, // 0 = automatisch (alle kernen)
    val autoStart: Boolean,
    val communityEnabled: Boolean,
    val hubUrl: String,
    val deviceId: String,
    val deviceName: String,
    val owner: String,
    val token: String,
) {
    companion object {
        private const val PREFS = "vhnw_config"

        fun load(context: Context): Config {
            val p = prefs(context)
            var deviceId = p.getString("device_id", "") ?: ""
            if (deviceId.isEmpty()) {
                deviceId = "vhnw-" + UUID.randomUUID().toString().replace("-", "").take(16)
                p.edit().putString("device_id", deviceId).apply()
            }
            return Config(
                pool = p.getString("pool", "rx.unmineable.com:3333")!!,
                algo = p.getString("algo", "rx")!!,
                tls = p.getBoolean("tls", false),
                wallet = p.getString("wallet", "")!!,
                worker = p.getString("worker", android.os.Build.MODEL ?: "android")!!,
                referral = p.getString("referral", "")!!,
                password = p.getString("password", "x")!!,
                threads = p.getInt("threads", 0),
                autoStart = p.getBoolean("auto_start", false),
                communityEnabled = p.getBoolean("community_enabled", false),
                hubUrl = p.getString("hub_url", "https://hash.vbnw.nl")!!,
                deviceId = deviceId,
                deviceName = p.getString("device_name", android.os.Build.MODEL ?: "android")!!,
                owner = p.getString("owner", "")!!,
                token = p.getString("token", "")!!,
            )
        }

        fun save(context: Context, config: Config) {
            prefs(context).edit().apply {
                putString("pool", config.pool)
                putString("algo", config.algo)
                putBoolean("tls", config.tls)
                putString("wallet", config.wallet)
                putString("worker", config.worker)
                putString("referral", config.referral)
                putString("password", config.password)
                putInt("threads", config.threads)
                putBoolean("auto_start", config.autoStart)
                putBoolean("community_enabled", config.communityEnabled)
                putString("hub_url", config.hubUrl)
                putString("device_name", config.deviceName)
                putString("owner", config.owner)
                putString("token", config.token)
                apply()
            }
        }

        private fun prefs(context: Context): SharedPreferences =
            context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)
    }
}
