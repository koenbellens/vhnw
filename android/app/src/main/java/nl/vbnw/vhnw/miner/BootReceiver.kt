package nl.vbnw.vhnw.miner

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent

/** Start het minen automatisch na een herstart, als dat in de instellingen aan staat. */
class BootReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action != Intent.ACTION_BOOT_COMPLETED) return
        val config = Config.load(context)
        if (config.autoStart) {
            MiningService.start(context)
        }
    }
}
