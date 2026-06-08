"use strict";

const $ = (id) => document.getElementById(id);

function fmtHr(v) {
  if (!v || v <= 0) return "0";
  const units = ["", "k", "M", "G", "T", "P"];
  let i = 0;
  while (v >= 1000 && i < units.length - 1) { v /= 1000; i++; }
  return (i === 0 ? v.toFixed(0) : v.toFixed(2)) + (units[i] ? " " + units[i] : "");
}

function fmtUptime(secs) {
  if (!secs || secs <= 0) return "—";
  const d = Math.floor(secs / 86400);
  const h = Math.floor((secs % 86400) / 3600);
  const m = Math.floor((secs % 3600) / 60);
  if (d > 0) return `${d}d ${h}u`;
  if (h > 0) return `${h}u ${m}m`;
  return `${m}m`;
}

function fmtSeen(secs) {
  if (secs < 60) return "zojuist";
  const m = Math.floor(secs / 60);
  if (m < 60) return `${m}m geleden`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}u geleden`;
  return `${Math.floor(h / 24)}d geleden`;
}

const TYPE_LABEL = { "cpu-agent": "CPU", "asic": "ASIC", "gpu": "GPU" };
const TYPE_CLASS = { "cpu-agent": "cpu", "asic": "asic", "gpu": "gpu" };

function esc(s) {
  return String(s == null ? "" : s).replace(/[&<>"]/g, (c) =>
    ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" }[c]));
}

function deviceCard(d) {
  const t = d.type || "cpu-agent";
  const badge = `<span class="badge ${TYPE_CLASS[t] || "cpu"}">${TYPE_LABEL[t] || t}</span>`;
  const dot = d.online ? "dot on" : "dot";
  const temp = d.temp_c ? d.temp_c.toFixed(1) + " °C" : "—";
  const state = d.online ? esc(d.state || "—") : "offline";
  const owner = d.owner ? `<div class="owner">van ${esc(d.owner)}</div>` : "";
  return `
    <div class="dev ${d.online ? "" : "off"}">
      <div class="top">
        <div class="name"><span class="${dot}"></span>${esc(d.name || "?")}</div>
        ${badge}
      </div>
      ${owner}
      <div class="hr">${fmtHr(d.hashrate)} <small>H/s</small></div>
      <div class="meta">
        <div class="row"><span>Status</span><span>${state}</span></div>
        <div class="row"><span>Shares</span><span>${d.shares_good || 0}</span></div>
        <div class="row"><span>Temp</span><span>${temp}</span></div>
        <div class="row"><span>Uptime</span><span>${fmtUptime(d.uptime_secs)}</span></div>
        <div class="row"><span>Herstarts</span><span>${d.restarts || 0}</span></div>
        <div class="row"><span>Gezien</span><span>${fmtSeen(d.last_seen_secs)}</span></div>
      </div>
      <div class="pool">${esc(d.pool || "—")} &middot; ${esc(d.miner || "—")}</div>
    </div>`;
}

async function poll() {
  try {
    const r = await fetch("/api/devices");
    const s = await r.json();
    $("live").textContent = "live";
    $("t_hr").textContent = fmtHr(s.total_hashrate);
    $("t_online").textContent = s.online_devices || 0;
    $("t_total").textContent = s.total_devices || 0;
    $("t_shares").textContent = s.total_shares || 0;

    const devices = s.devices || [];
    if (devices.length === 0) {
      $("devices").innerHTML = "";
      $("empty").style.display = "block";
    } else {
      $("empty").style.display = "none";
      $("devices").innerHTML = devices.map(deviceCard).join("");
    }
  } catch (e) {
    $("live").textContent = "geen verbinding";
  }
}

poll();
setInterval(poll, 3000);
