"use strict";

const $ = (id) => document.getElementById(id);

let desired = false;     // wat de gebruiker wil (laatst bekende toestand)
let busy = false;        // voorkomt dubbele klikken
let cfgLoaded = false;

function fmtHr(v) {
  if (!v || v <= 0) return "0";
  if (v >= 1000) return (v / 1000).toFixed(2) + " k";
  return v.toFixed(1);
}

async function poll() {
  try {
    const r = await fetch("/api/status");
    const s = await r.json();
    render(s);
  } catch (e) {
    $("state").textContent = "geen verbinding";
    $("dot").className = "dot err";
  }
}

function render(s) {
  const m = s.miner || {};
  desired = !!m.desired;

  // statusbadge
  let cls = "dot", label = m.state || "onbekend";
  if (m.state === "actief") cls = "dot run";
  else if (m.state === "fout") cls = "dot err";
  else if (m.state === "herstarten" || m.state === "bezig met starten") cls = "dot wait";
  $("dot").className = cls;
  $("state").textContent = label;

  // power-knop
  const btn = $("power");
  btn.disabled = busy || (!m.config_valid && !desired);
  if (desired) {
    btn.classList.add("on");
    btn.textContent = "STOP";
  } else {
    btn.classList.remove("on");
    btn.textContent = "START";
  }

  // config-waarschuwing
  if (!m.config_valid && m.config_error) {
    $("cfgwarn").style.display = "block";
    $("cfgwarn").textContent = "Vul eerst je instellingen in: " + m.config_error;
  } else {
    $("cfgwarn").style.display = "none";
  }

  // stats
  const sum = s.summary;
  $("hr").textContent = fmtHr(sum ? sum.hashrate_now : 0);
  $("hr15").textContent = fmtHr(sum ? sum.hashrate_15m : 0);
  $("shares").textContent = sum ? sum.shares_good : 0;
  $("acc").textContent = sum ? sum.accepted : 0;
  $("rej").textContent = sum ? sum.rejected : 0;
  $("ver").textContent = sum && sum.version ? sum.version : "—";
  $("pool").textContent = sum && sum.pool ? sum.pool : "—";
  $("uptime").textContent = m.running_secs ? fmtUptime(m.running_secs) : "—";
  $("restarts").textContent = m.restart_count || 0;

  const sys = s.system || {};
  $("host").textContent = sys.hostname || "—";
  $("temp").textContent = sys.temp_c ? sys.temp_c.toFixed(1) : "—";
}

function fmtUptime(secs) {
  const d = Math.floor(secs / 86400);
  const h = Math.floor((secs % 86400) / 3600);
  const m = Math.floor((secs % 3600) / 60);
  if (d > 0) return `${d}d ${h}u ${m}m`;
  if (h > 0) return `${h}u ${m}m`;
  return `${m}m`;
}

async function pollLogs() {
  try {
    const r = await fetch("/api/logs");
    const data = await r.json();
    const el = $("logs");
    const atBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 20;
    el.textContent = (data.lines || []).join("\n") || "(nog geen log)";
    if (atBottom) el.scrollTop = el.scrollHeight;
  } catch (e) { /* stil */ }
}

$("power").addEventListener("click", async () => {
  if (busy) return;
  busy = true;
  const target = desired ? "/api/stop" : "/api/start";
  try {
    const r = await fetch(target, { method: "POST" });
    const data = await r.json();
    if (data.error) alert(data.error);
  } catch (e) {
    alert("Kon de actie niet uitvoeren");
  } finally {
    busy = false;
    poll();
  }
});

async function loadConfig() {
  const r = await fetch("/api/config");
  const c = await r.json();
  $("f_wallet").value = c.wallet || "";
  $("f_worker").value = c.worker || "";
  $("f_pool").value = c.pool || "";
  $("f_algo").value = c.algo || "";
  $("f_pass").value = c.password || "";
  $("f_ref").value = c.referral || "";
  $("f_threads").value = c.threads || 0;
  $("f_tls").value = c.tls ? "true" : "false";
  $("f_auto").value = c.auto_start ? "true" : "false";
  cfgLoaded = true;
}

$("save").addEventListener("click", async () => {
  const msg = $("savemsg");
  const body = {
    wallet: $("f_wallet").value.trim(),
    worker: $("f_worker").value.trim(),
    pool: $("f_pool").value.trim(),
    algo: $("f_algo").value.trim(),
    password: $("f_pass").value,
    referral: $("f_ref").value.trim(),
    threads: parseInt($("f_threads").value || "0", 10),
    tls: $("f_tls").value === "true",
    auto_start: $("f_auto").value === "true",
  };
  try {
    const r = await fetch("/api/config", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const data = await r.json();
    if (data.error) {
      msg.className = "msg err"; msg.textContent = data.error;
    } else {
      msg.className = "msg ok"; msg.textContent = "opgeslagen";
      setTimeout(() => (msg.textContent = ""), 3000);
    }
  } catch (e) {
    msg.className = "msg err"; msg.textContent = "opslaan mislukt";
  }
});

loadConfig();
poll();
pollLogs();
setInterval(poll, 2000);
setInterval(pollLogs, 3000);
