# ============================================
# RB951 Hotspot + RADIUS + Fairness Baseline
# Adjust placeholders before import
# ============================================

# ---------- 0) Variables ----------
:local WAN_IF "ether1"
:local LAN_BRIDGE "bridge"
:local HOTSPOT_SERVER "hotspot1"
:local HOTSPOT_PROFILE "hsprof1"
:local RADIUS_IP "192.168.88.2"
:local RADIUS_SECRET "CHANGE_ME_RADIUS_SECRET"
:local COA_SECRET "change-me"

# ---------- 1) Enable RADIUS for Hotspot ----------
/radius
add address=$RADIUS_IP secret=$RADIUS_SECRET service=hotspot timeout=300ms authentication-port=1812 accounting-port=1813

/radius incoming
set accept=yes port=3799

# ---------- 2) Router identity in RADIUS NAS table ----------
# In your radius `nas` table add:
# nasname=<RB951_LAN_IP>, shortname=rb951, secret=<same as RADIUS_SECRET>

# ---------- 3) Hotspot profile tweaks ----------
/ip hotspot profile
set [find name=$HOTSPOT_PROFILE] use-radius=yes radius-accounting=yes interim-update=1m login-by=http-chap,http-pap

# ---------- 4) Optional local profile (fallback if RADIUS is down) ----------
/ip hotspot user profile
add name="4Mbps_Plan_Local" shared-users=1 rate-limit="4M/4M" status-autorefresh=1m \
    on-login=":log info (\"Hotspot user \$user logged in (local 4M profile)\")"

# ---------- 5) PCQ fairness (equal-share under congestion) ----------
# This gives per-user fairness when backhaul is saturated.
/queue type
add name=pcq-download-4m kind=pcq pcq-classifier=dst-address pcq-rate=4M pcq-total-limit=2000KiB
add name=pcq-upload-4m kind=pcq pcq-classifier=src-address pcq-rate=4M pcq-total-limit=2000KiB

# Parent queue for total Airtel pipe (change to your real backhaul, e.g. 30M/30M)
/queue tree
add name="TOTAL-DOWNLINK" parent=global max-limit=30M queue=default
add name="TOTAL-UPLINK" parent=global max-limit=30M queue=default
add name="FAIR-DOWNLINK" parent="TOTAL-DOWNLINK" packet-mark="" queue=pcq-download-4m
add name="FAIR-UPLINK" parent="TOTAL-UPLINK" packet-mark="" queue=pcq-upload-4m

# ---------- 6) Optional local expiry fallback ----------
# Preferred expiry is already in your backend + RADIUS Session-Timeout + CoA.
# Keep this only as fallback logic for locally-created hotspot users.
/system script
add name=expire-local-hotspot-users policy=read,write,test source={
  :local now [/system clock get time];
  :foreach i in=[/ip hotspot active find] do={
    :local user [/ip hotspot active get $i user];
    :local uptime [/ip hotspot active get $i uptime];
    # Example fallback: remove session after 1h if username starts with "tmp-"
    :if ([:pick $user 0 4] = "tmp-") do={
      :if ($uptime > 01:00:00) do={
        /ip hotspot active remove $i;
        :log warning ("Expired local tmp user: ".$user);
      }
    }
  }
}

/system scheduler
add name=run-expire-local-hotspot-users interval=1m on-event=expire-local-hotspot-users

# ---------- 7) Walled garden + login page baseline ----------
# Replace with your billing portal and payment callback landing pages.
/ip hotspot walled-garden
add dst-host="your-billing-domain.com"
add dst-host="*.safaricom.co.ke"

# ---------- 8) Verification ----------
# /radius print detail
# /ip hotspot profile print detail
# /ip hotspot active print
# /queue tree print stats

