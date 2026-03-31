import { useEffect, useMemo, useState } from "react";
import "./App.css";

const packages = [
  { duration: "30 mins", price: "Ksh. 7", amount: 7, tag: "4 Mbps", dataCapMB: 300 },
  { duration: "1 hour", price: "Ksh. 10", amount: 10, tag: "4 Mbps", dataCapMB: 500 },
  { duration: "2 hours", price: "Ksh. 15", amount: 15, tag: "4 Mbps", dataCapMB: 1024 },
  { duration: "6 hours", price: "Ksh. 20", amount: 20, tag: "4 Mbps", dataCapMB: 2048 },
  { duration: "12 hours", price: "Ksh. 25", amount: 25, tag: "4 Mbps", dataCapMB: 3072 },
  { duration: "24 hours", price: "Ksh. 30", amount: 30, tag: "4 Mbps", dataCapMB: 4096 },
  { duration: "3 days", price: "Ksh. 100", amount: 100, tag: "4 Mbps", dataCapMB: 8192 },
  { duration: "7 days", price: "Ksh. 150", amount: 150, tag: "5 Mbps", dataCapMB: 15360 },
  { duration: "15 days", price: "Ksh. 250", amount: 250, tag: "5 Mbps", dataCapMB: 30720 },
  { duration: "30 days", price: "Ksh. 350", amount: 350, tag: "5 Mbps", dataCapMB: 51200 },
  { duration: "30 days", price: "Ksh. 500", amount: 500, tag: "8 Mbps", dataCapMB: 81920 },
  { duration: "30 days", price: "Ksh. 800", amount: 800, tag: "10 Mbps", dataCapMB: 122880 }
];

const subIspPackages = [
  { name: "Lite plan", price: "Ksh. 500", amount: 500, maxUsers: 50, maxRouters: 2, duration: "30 days" },
  { name: "Pro Plan", price: "Ksh. 1,250", amount: 1250, maxUsers: 150, maxRouters: 5, duration: "30 days" },
  { name: "Elite Plan", price: "Ksh. 2,000", amount: 2000, maxUsers: 400, maxRouters: 10, duration: "30 days" },
  { name: "Unlimited Plan", price: "Ksh. 2,500", amount: 2500, maxUsers: -1, maxRouters: -1, duration: "30 days" }
];

export default function App() {
  const apiBase = import.meta.env.VITE_API_URL || "";
  const adminPath = (import.meta.env.VITE_ADMIN_PATH || "/admin").trim();
  const adminHost = (import.meta.env.VITE_ADMIN_HOST || "").trim().toLowerCase();
  const locationPath = typeof window !== "undefined" ? window.location.pathname : "/";
  const locationHost = typeof window !== "undefined" ? window.location.hostname.toLowerCase() : "";
  const onAdminPath = adminPath === "/" ? true : locationPath === adminPath || locationPath.startsWith(adminPath + "/");
  const onAdminHost = adminHost !== "" && locationHost === adminHost;
  const authEnabled = adminHost !== "" ? onAdminHost : onAdminPath;
  const [selected, setSelected] = useState(null);
  const [phone, setPhone] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [sending, setSending] = useState(false);
  const [status, setStatus] = useState("");
  const [authOpen, setAuthOpen] = useState(false);
  const [authEmail, setAuthEmail] = useState("");
  const [authPassword, setAuthPassword] = useState("");
  const [authError, setAuthError] = useState("");
  const [token, setToken] = useState(localStorage.getItem("tst_token") || "");
  const [role, setRole] = useState(localStorage.getItem("tst_role") || "");
  const [view, setView] = useState("billing");
  const [users, setUsers] = useState([]);
  const [revenue, setRevenue] = useState({ items: [], total: 0, count: 0 });
  const [adminLoading, setAdminLoading] = useState(false);
  const [adminError, setAdminError] = useState("");
  const [usageLoading, setUsageLoading] = useState(false);
  const [usageError, setUsageError] = useState("");
  const [usageMap, setUsageMap] = useState({});
  const [userFilter, setUserFilter] = useState("all");
  const [adminTab, setAdminTab] = useState("overview");
  const [subRegOpen, setSubRegOpen] = useState(false);
  const [subRegPackage, setSubRegPackage] = useState(null);
  const [subRegData, setSubRegData] = useState({
    business: "",
    contact: "",
    email: "",
    password: "",
    phone: "",
    location: "",
  });
  const [subRegSubmitted, setSubRegSubmitted] = useState(false);
  const [subRegSending, setSubRegSending] = useState(false);
  const [subRegStatus, setSubRegStatus] = useState("");
  const [subIspFilter, setSubIspFilter] = useState("all");
  const [subIsps, setSubIsps] = useState([]);
  const [subIspLoading, setSubIspLoading] = useState(false);
  const [subIspError, setSubIspError] = useState("");
  const [subProfile, setSubProfile] = useState(null);
  const [subProfileLoading, setSubProfileLoading] = useState(false);
  const [subProfileError, setSubProfileError] = useState("");
  const [routerName, setRouterName] = useState("");
  const [routers, setRouters] = useState([]);
  const [ownerRouters, setOwnerRouters] = useState([]);
  const [ownerRouterForm, setOwnerRouterForm] = useState({
    name: "",
    host: "",
    secret: "",
    serviceType: "pppoe",
    coaPort: 3799,
    authPort: 1812,
    acctPort: 1813,
    nasId: "",
    enabled: true,
  });
  const [ownerRouterStatus, setOwnerRouterStatus] = useState("");
  const [showRouterForm, setShowRouterForm] = useState(false);
  const [ownerRouterTests, setOwnerRouterTests] = useState({});

  const phoneError = useMemo(() => {
    if (!phone) return "";
    if (!phone.startsWith("07") && !phone.startsWith("01")) return "Phone must start with 07 or 01.";
    if (phone.length !== 10) return "Phone must be 10 digits.";
    if (!/^\d+$/.test(phone)) return "Phone must be numeric.";
    return "";
  }, [phone]);

  const subPhoneError = useMemo(() => {
    if (!subRegData.phone) return "";
    if (!subRegData.phone.startsWith("07") && !subRegData.phone.startsWith("01")) return "Phone must start with 07 or 01.";
    if (subRegData.phone.length !== 10) return "Phone must be 10 digits.";
    if (!/^\d+$/.test(subRegData.phone)) return "Phone must be numeric.";
    return "";
  }, [subRegData.phone]);
  const subPasswordError = useMemo(() => {
    if (!subRegData.password) return "";
    if (subRegData.password.length < 8) return "Password must be at least 8 characters.";
    return "";
  }, [subRegData.password]);

  const canPay = selected && !phoneError && phone;
  const canSubmitSubReg =
    subRegPackage &&
    subRegData.business &&
    subRegData.email &&
    subRegData.phone &&
    !subPhoneError &&
    !subPasswordError;
  const isSubIsp = token && role === "sub_isp";
  const isOwner = authEnabled && token && !isSubIsp;
  const isSubBilling = view === "sub_billing";
  const filteredSubIsps = subIsps.filter((sub) => {
    if (subIspFilter === "all") return true;
    return sub.status === subIspFilter;
  });
  const subStatus = subProfile?.status || "";
  const subStatusLabel = subStatus === "active" ? "Active" : subStatus === "suspended" ? "Suspended" : "Pending";
  const subRouterLimit = typeof subProfile?.maxRouters === "number" ? subProfile.maxRouters : -1;
  const subRouterCount = typeof subProfile?.routerCount === "number" ? subProfile.routerCount : routers.length;
  const subRouterUnlimited = subRouterLimit < 0;
  const subRouterSlotsLeft = subRouterUnlimited ? Infinity : Math.max(subRouterLimit - subRouterCount, 0);
  const canAddSubRouter = subStatus === "active" && routerName.trim() && (subRouterUnlimited || subRouterSlotsLeft > 0);

  const formatBytes = (bytes) => {
    if (!bytes || bytes <= 0) return "0 B";
    const units = ["B", "KB", "MB", "GB", "TB"];
    let value = bytes;
    let i = 0;
    while (value >= 1024 && i < units.length - 1) {
      value /= 1024;
      i += 1;
    }
    return value.toFixed(value >= 10 ? 0 : 1) + " " + units[i];
  };

  const packageCapBytes = (pkgName) => {
    const pkg = packages.find((p) => `${p.duration} ${p.tag}` === pkgName);
    if (!pkg || !pkg.dataCapMB) return 0;
    return pkg.dataCapMB * 1024 * 1024;
  };

  const formatRemaining = (pkgName, usageBytes) => {
    const cap = packageCapBytes(pkgName);
    if (!cap) return "Unlimited";
    const remaining = Math.max(cap - (usageBytes || 0), 0);
    return formatBytes(remaining);
  };

  const formatRemainingTime = (paidUntil) => {
    if (!paidUntil) return "-";
    const remainingMs = new Date(paidUntil).getTime() - Date.now();
    if (remainingMs <= 0) return "Expired";
    return formatSeconds(Math.floor(remainingMs / 1000));
  };
  const formatSeconds = (seconds) => {
    if (!seconds || seconds <= 0) return "0s";
    const hrs = Math.floor(seconds / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    if (hrs > 0) return hrs + "h " + mins + "m";
    return mins + "m";
  };

  function openPay(pkg) {
    setSelected(pkg);
    setPhone("");
    setSubmitted(false);
    setSending(false);
    setStatus("");
  }

  function closePay() {
    setSelected(null);
  }

  function openSubReg(pkg) {
    setSubRegOpen(true);
    setSubRegSubmitted(false);
    setSubRegSending(false);
    setSubRegStatus("");
    setSubRegData({
      business: "",
      contact: "",
      email: "",
      password: "",
      phone: "",
      location: "",
    });
    setSubRegPackage(pkg || null);
  }

  function closeSubReg() {
    setSubRegOpen(false);
  }

  function openSubIspLogin() {
    closeSubReg();
    openLogin();
  }

  function submitPay(e) {
    e.preventDefault();
    setSubmitted(true);
    if (!canPay) return;
    setSending(true);
    setStatus("");

    fetch(apiBase + "/api/mpesa/stkpush", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        phone,
        amount: selected.amount,
        packageName: selected.duration + " " + selected.tag
      })
    })
      .then(async (res) => {
        const data = await res.json().catch(() => ({}));
        if (!res.ok) {
          throw new Error(data?.message || "Failed to send prompt.");
        }
        return data;
      })
      .then((data) => {
        setStatus(data?.message || "Prompt sent. Check your phone to complete payment.");
      })
      .catch((err) => {
        setStatus(err.message || "Something went wrong.");
      })
      .finally(() => {
        setSending(false);
      });
  }

  function submitSubReg(e) {
    e.preventDefault();
    setSubRegSubmitted(true);
    if (!canSubmitSubReg) return;
    setSubRegSending(true);
    setSubRegStatus("");

    fetch(apiBase + "/api/subisp/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        business: subRegData.business,
        contact: subRegData.contact,
        email: subRegData.email,
        password: subRegData.password,
        phone: subRegData.phone,
        location: subRegData.location,
        packageName: subRegPackage.name,
        amount: subRegPackage.amount
      })
    })
      .then(async (res) => {
        const data = await res.json().catch(() => ({}));
        if (!res.ok) {
          throw new Error(data?.message || "Failed to send prompt.");
        }
        return data;
      })
      .then((data) => {
        const banner =
          "Payment required. Your Sub-ISP account is pending until you complete Mpesa payment. Log in after payment to access the dashboard.";
        setSubRegStatus(data?.message || banner);
        setSubRegSubmitted(false);
      })
      .catch((err) => {
        setSubRegStatus(err.message || "Something went wrong.");
      })
      .finally(() => {
        setSubRegSending(false);
      });
  }

  function openLogin() {
    setAuthOpen(true);
    setAuthError("");
    setAuthPassword("");
  }

  function closeLogin() {
    setAuthOpen(false);
  }

  function logout() {
    localStorage.removeItem("tst_token");
    localStorage.removeItem("tst_role");
    setToken("");
    setRole("");
    setView("billing");
  }

  function submitLogin(e) {
    e.preventDefault();
    setAuthError("");
    fetch(apiBase + "/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: authEmail, password: authPassword })
    })
      .then(async (res) => {
        const data = await res.json().catch(() => ({}));
        if (!res.ok) {
          throw new Error(data?.message || "Login failed.");
        }
        return data;
      })
      .then((data) => {
        localStorage.setItem("tst_token", data.token);
        localStorage.setItem("tst_role", data.role);
        setToken(data.token);
        setRole(data.role);
        setView(data.role === "sub_isp" ? "sub_dashboard" : "owner_dashboard");
        setAuthOpen(false);
      })
      .catch((err) => {
        setAuthError(err.message || "Login failed.");
      });
  }

  function updateSubIspPlan(id, plan) {
    const prev = subIsps;
    setSubIsps((items) => items.map((sub) => (sub.id === id ? { ...sub, plan } : sub)));
    setSubIspError("");
    fetch(apiBase + "/api/admin/subisps/update", {
      method: "PATCH",
      headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
      body: JSON.stringify({ id, plan })
    })
      .then(async (res) => {
        const data = await res.json().catch(() => ({}));
        if (!res.ok) throw new Error(data?.message || "Failed to update plan.");
        return data;
      })
      .then((data) => {
        if (data && data.id) {
          setSubIsps((items) => items.map((sub) => (sub.id === data.id ? data : sub)));
        }
      })
      .catch((err) => {
        setSubIsps(prev);
        setSubIspError(err.message || "Failed to update plan.");
      });
  }

  function toggleSubIspStatus(id) {
    const target = subIsps.find((sub) => sub.id === id);
    if (!target) return;
    const nextStatus = target.status === "active" ? "suspended" : "active";
    const prev = subIsps;
    setSubIsps((items) => items.map((sub) => (sub.id === id ? { ...sub, status: nextStatus } : sub)));
    setSubIspError("");
    fetch(apiBase + "/api/admin/subisps/update", {
      method: "PATCH",
      headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
      body: JSON.stringify({ id, status: nextStatus })
    })
      .then(async (res) => {
        const data = await res.json().catch(() => ({}));
        if (!res.ok) throw new Error(data?.message || "Failed to update status.");
        return data;
      })
      .then((data) => {
        if (data && data.id) {
          setSubIsps((items) => items.map((sub) => (sub.id === data.id ? data : sub)));
        }
      })
      .catch((err) => {
        setSubIsps(prev);
        setSubIspError(err.message || "Failed to update status.");
      });
  }

  function addRouter() {
    const trimmed = routerName.trim();
    if (!trimmed) return;
    if (subStatus !== "active") {
      setSubProfileError("Your account is not active. Activate your plan to add routers.");
      return;
    }
    if (!subRouterUnlimited && subRouterSlotsLeft <= 0) {
      setSubProfileError("Router limit reached for your current plan.");
      return;
    }
    setSubProfileError("");
    fetch(apiBase + "/api/subisp/routers", {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
      body: JSON.stringify({ name: trimmed })
    })
      .then(async (res) => {
        const data = await res.json().catch(() => ({}));
        if (!res.ok) throw new Error(data?.message || "Failed to add router.");
        return data;
      })
      .then((data) => {
        if (data && data.id) {
          setRouters((prev) => [...prev, data]);
          setSubProfile((prev) =>
            prev
              ? {
                  ...prev,
                  routerCount: (typeof prev.routerCount === "number" ? prev.routerCount : 0) + 1
                }
              : prev
          );
        }
        setRouterName("");
      })
      .catch((err) => {
        setSubProfileError(err.message || "Failed to add router.");
      });
  }

  function updateRouterStatus(id, status) {
    setSubProfileError("");
    fetch(apiBase + "/api/subisp/routers", {
      method: "PUT",
      headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
      body: JSON.stringify({ id, status })
    })
      .then(async (res) => {
        const data = await res.json().catch(() => ({}));
        if (!res.ok) throw new Error(data?.message || "Failed to update router.");
        return data;
      })
      .then((data) => {
        if (data && data.id) {
          setRouters((prev) => prev.map((r) => (r.id === data.id ? data : r)));
        }
      })
      .catch((err) => {
        setSubProfileError(err.message || "Failed to update router.");
      });
  }

  function updateOwnerRouterField(key, value) {
    setOwnerRouterForm((prev) => ({ ...prev, [key]: value }));
  }

  function submitOwnerRouter(e) {
    e.preventDefault();
    setOwnerRouterStatus("");
    fetch(apiBase + "/api/admin/routers", {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
      body: JSON.stringify(ownerRouterForm)
    })
      .then(async (res) => {
        const data = await res.json().catch(() => ({}));
        if (!res.ok) throw new Error(data?.message || "Failed to add router.");
        return data;
      })
      .then(() => {
        setOwnerRouterStatus("Router saved.");
        setOwnerRouterForm({
          name: "",
          host: "",
          secret: "",
          serviceType: "pppoe",
          coaPort: 3799,
          authPort: 1812,
          acctPort: 1813,
          nasId: "",
          enabled: true,
        });
        return fetch(apiBase + "/api/admin/routers", {
          headers: { Authorization: "Bearer " + token }
        }).then((res) => res.json());
      })
      .then((data) => {
        setOwnerRouters(Array.isArray(data) ? data : []);
        setShowRouterForm(false);
      })
      .catch((err) => {
        setOwnerRouterStatus(err.message || "Failed to add router.");
      });
  }

  function deleteOwnerRouter(id) {
    if (!id) return;
    fetch(apiBase + "/api/admin/routers?id=" + id, {
      method: "DELETE",
      headers: { Authorization: "Bearer " + token }
    })
      .then(() => {
        setOwnerRouters((prev) => prev.filter((r) => r.id !== id));
      })
      .catch(() => {
        setOwnerRouterStatus("Failed to delete router.");
      });
  }

  function testOwnerRouter(id) {
    if (!id) return;
    setOwnerRouterTests((prev) => ({ ...prev, [id]: "Testing..." }));
    fetch(apiBase + "/api/admin/routers/test?id=" + id, {
      method: "POST",
      headers: { Authorization: "Bearer " + token }
    })
      .then(async (res) => {
        const data = await res.json().catch(() => ({}));
        if (!res.ok) throw new Error(data?.message || "Test failed.");
        return data;
      })
      .then((data) => {
        setOwnerRouterTests((prev) => ({ ...prev, [id]: data?.message || "OK" }));
      })
      .catch((err) => {
        setOwnerRouterTests((prev) => ({ ...prev, [id]: err.message || "Test failed." }));
      });
  }

  function fetchUsage(username) {
    if (!username) return;
    setUsageLoading(true);
    setUsageError("");
    fetch(apiBase + "/api/admin/usage?username=" + encodeURIComponent(username), {
      headers: { Authorization: "Bearer " + token }
    })
      .then((res) => res.json())
      .then((data) => {
        if (data && data.username) {
          setUsageMap((prev) => ({
            ...prev,
            [data.username]: {
              bytesUsed: data.bytesUsed || 0,
              timeUsed: data.timeUsedSeconds || 0
            }
          }));
        }
      })
      .catch(() => setUsageError("Failed to load usage."))
      .finally(() => setUsageLoading(false));
  }

  const [subUsageQuery, setSubUsageQuery] = useState("");
  const [subUsageLoading, setSubUsageLoading] = useState(false);
  const [subUsageError, setSubUsageError] = useState("");
  const [subUsageResult, setSubUsageResult] = useState(null);

  function fetchSubUsage() {
    const username = subUsageQuery.trim();
    if (!username) return;
    setSubUsageLoading(true);
    setSubUsageError("");
    setSubUsageResult(null);
    fetch(apiBase + "/api/subisp/usage?username=" + encodeURIComponent(username), {
      headers: { Authorization: "Bearer " + token }
    })
      .then((res) => res.json())
      .then((data) => {
        if (data && data.username) {
          setSubUsageResult(data);
        }
      })
      .catch(() => setSubUsageError("Failed to load usage."))
      .finally(() => setSubUsageLoading(false));
  }

  useEffect(() => {
    if (!authEnabled) return;
    if (!token) return;
    if (isSubIsp) return;
    setAdminLoading(true);
    setAdminError("");
    const statusParam = userFilter === "all" ? "" : "?status=" + userFilter;
    Promise.all([
      fetch(apiBase + "/api/admin/users" + statusParam, {
        headers: { Authorization: "Bearer " + token }
      }).then((res) => res.json()),
      fetch(apiBase + "/api/admin/revenue", {
        headers: { Authorization: "Bearer " + token }
      }).then((res) => res.json()),
      fetch(apiBase + "/api/admin/routers", {
        headers: { Authorization: "Bearer " + token }
      }).then((res) => res.json())
    ])
      .then(([usersData, revenueData, routersData]) => {
        setUsers(Array.isArray(usersData) ? usersData : []);
        setRevenue(revenueData && revenueData.items ? revenueData : { items: [], total: 0, count: 0 });
        setOwnerRouters(Array.isArray(routersData) ? routersData : []);
      })
      .catch(() => {
        setAdminError("Failed to load admin data.");
      })
      .finally(() => setAdminLoading(false));
  }, [token, apiBase, userFilter, isSubIsp, authEnabled]);

  useEffect(() => {
    if (!token || !isOwner) return;
    setSubIspLoading(true);
    setSubIspError("");
    const statusParam = subIspFilter === "all" ? "" : "?status=" + subIspFilter;
    fetch(apiBase + "/api/admin/subisps" + statusParam, {
      headers: { Authorization: "Bearer " + token }
    })
      .then((res) => res.json())
      .then((data) => {
        setSubIsps(Array.isArray(data) ? data : []);
      })
      .catch(() => {
        setSubIspError("Failed to load Sub-ISPs.");
      })
      .finally(() => setSubIspLoading(false));
  }, [token, apiBase, isOwner, subIspFilter]);

  useEffect(() => {
    if (!token || !isSubIsp) return;
    setSubProfileLoading(true);
    setSubProfileError("");
    fetch(apiBase + "/api/subisp/me", {
      headers: { Authorization: "Bearer " + token }
    })
      .then((res) => res.json())
      .then((data) => {
        setSubProfile(data || null);
        setRouters(Array.isArray(data?.routers) ? data.routers : []);
      })
      .catch(() => {
        setSubProfileError("Failed to load Sub-ISP profile.");
      })
      .finally(() => setSubProfileLoading(false));
  }, [token, apiBase, isSubIsp]);

  return (
    <main className="page">
      <div className="topbar">
        <div className="brand">
          <span className="brand-mark">TST-ISP</span>
          {authEnabled && <span className="brand-sub">{isSubIsp ? "Sub-ISP Portal" : "Owner Portal"}</span>}
        </div>
        <div className="top-actions">
          {authEnabled ? (
            token ? (
              <>
                {!isSubIsp && (
                  <button className={view === "billing" ? "ghost active" : "ghost"} onClick={() => setView("billing")}>
                    Billing
                  </button>
                )}
                <button
                  className={view === (isSubIsp ? "sub_dashboard" : "owner_dashboard") ? "ghost active" : "ghost"}
                  onClick={() => setView(isSubIsp ? "sub_dashboard" : "owner_dashboard")}
                >
                  Dashboard
                </button>
                <button className="login-btn" onClick={logout}>
                  Logout
                </button>
              </>
            ) : (
              <button className="login-btn" onClick={openLogin}>
                Login
              </button>
            )
          ) : isSubIsp ? (
            <>
              <button
                className={view === "sub_dashboard" ? "ghost active" : "ghost"}
                onClick={() => setView("sub_dashboard")}
              >
                Dashboard
              </button>
              <button className="login-btn" onClick={logout}>
                Logout
              </button>
            </>
          ) : null}
        </div>
      </div>

      {authEnabled && view === "owner_dashboard" && token && isOwner ? (
        <section className="dashboard">
          <header className="dash-header">
            <div>
              <p className="kicker">Super Admin</p>
              <h2 className="dash-title">Owner Dashboard</h2>
              <p className="subtitle">Full access to users, activity status, revenue, routers, and Sub-ISP operations.</p>
            </div>
            <div className="dash-role">{role || "super_admin"}</div>
          </header>

          <div className="filters">
            <button className={adminTab === "overview" ? "ghost active" : "ghost"} onClick={() => setAdminTab("overview")}>
              Overview
            </button>
            <button className={adminTab === "subisps" ? "ghost active" : "ghost"} onClick={() => setAdminTab("subisps")}>
              Sub-ISPs
            </button>
            <button className={adminTab === "users" ? "ghost active" : "ghost"} onClick={() => setAdminTab("users")}>
              Users
            </button>
          </div>

          {adminLoading && <p className="status">Loading admin data...</p>}
          {adminError && <p className="error">{adminError}</p>}
          {subIspError && <p className="error">{subIspError}</p>}
          {usageError && <p className="error">{usageError}</p>}

          <div className={adminTab === "overview" ? "dash-grid overview-grid" : "dash-grid"}>
            {adminTab === "overview" && (
              <section className="dash-card dash-wide">
              <div className="router-header">
                <h3>MikroTik Routers</h3>
                <button className="ghost" onClick={() => setShowRouterForm((prev) => !prev)}>
                  {showRouterForm ? "Hide Form" : "Add Router"}
                </button>
              </div>

              {ownerRouterStatus && <p className="status">{ownerRouterStatus}</p>}

              {showRouterForm && (
                <form className="router-form" onSubmit={submitOwnerRouter}>
                  <div className="dist-grid">
                    <label className="field">
                      <span>Router Name</span>
                      <input
                        className="input"
                        value={ownerRouterForm.name}
                        onChange={(e) => updateOwnerRouterField("name", e.target.value)}
                        placeholder="HQ Router"
                        required
                      />
                    </label>
                    <label className="field">
                      <span>Host / IP</span>
                      <input
                        className="input"
                        value={ownerRouterForm.host}
                        onChange={(e) => updateOwnerRouterField("host", e.target.value)}
                        placeholder="192.168.88.1"
                        required
                      />
                    </label>
                    <label className="field">
                      <span>Shared Secret</span>
                      <input
                        className="input"
                        value={ownerRouterForm.secret}
                        onChange={(e) => updateOwnerRouterField("secret", e.target.value)}
                        placeholder="radius-secret"
                        required
                      />
                    </label>
                  </div>

                  <div className="dist-grid">
                    <label className="field">
                      <span>Service Type</span>
                      <select
                        className="select"
                        value={ownerRouterForm.serviceType}
                        onChange={(e) => updateOwnerRouterField("serviceType", e.target.value)}
                      >
                        <option value="pppoe">PPPoE</option>
                        <option value="hotspot">Hotspot</option>
                      </select>
                    </label>
                    <label className="field">
                      <span>NAS Identifier</span>
                      <input
                        className="input"
                        value={ownerRouterForm.nasId}
                        onChange={(e) => updateOwnerRouterField("nasId", e.target.value)}
                        placeholder="main-nas"
                      />
                    </label>
                    <label className="field">
                      <span>Enabled</span>
                      <select
                        className="select"
                        value={ownerRouterForm.enabled ? "true" : "false"}
                        onChange={(e) => updateOwnerRouterField("enabled", e.target.value === "true")}
                      >
                        <option value="true">Yes</option>
                        <option value="false">No</option>
                      </select>
                    </label>
                  </div>

                  <div className="router-ports">
                    <label className="field">
                      <span>CoA Port</span>
                      <input
                        className="input"
                        type="number"
                        value={ownerRouterForm.coaPort}
                        onChange={(e) => updateOwnerRouterField("coaPort", Number(e.target.value) || 0)}
                      />
                    </label>
                    <label className="field">
                      <span>Auth Port</span>
                      <input
                        className="input"
                        type="number"
                        value={ownerRouterForm.authPort}
                        onChange={(e) => updateOwnerRouterField("authPort", Number(e.target.value) || 0)}
                      />
                    </label>
                    <label className="field">
                      <span>Acct Port</span>
                      <input
                        className="input"
                        type="number"
                        value={ownerRouterForm.acctPort}
                        onChange={(e) => updateOwnerRouterField("acctPort", Number(e.target.value) || 0)}
                      />
                    </label>
                  </div>

                  <button className="pay-btn" type="submit">
                    Save Router
                  </button>
                </form>
              )}

              <div className="router-list">
                {ownerRouters.length === 0 ? (
                  <p className="muted">No routers added yet.</p>
                ) : (
                  ownerRouters.map((router) => (
                    <div key={router.id || router.host} className="router-row">
                      <div>
                        <strong>{router.name || "Unnamed router"}</strong>
                        <p className="muted">
                          {router.host || "-"} | {router.serviceType || "pppoe"} |{" "}
                          {router.enabled === false ? "Disabled" : "Enabled"}
                        </p>
                      </div>
                      <div className="router-meta">
                        <button className="ghost small" onClick={() => testOwnerRouter(router.id)}>
                          Test
                        </button>
                        <button className="ghost small" onClick={() => deleteOwnerRouter(router.id)}>
                          Delete
                        </button>
                        <span className="muted">{ownerRouterTests[router.id] || "-"}</span>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </section>
            )}

            {adminTab === "overview" && (
              <section className="dash-card compact">
              <h3>Revenue by Package</h3>
              <div className="revenue-total">
                <span>Total Revenue</span>
                <strong>Ksh. {revenue.total}</strong>
                <span className="muted">{revenue.count} payments</span>
              </div>
              {revenue.items.length === 0 && !adminLoading && <p className="muted">No revenue yet.</p>}
              {revenue.items.map((row) => (
                <div key={row.package} className="revenue-row">
                  <span className="revenue-name">{row.package || "Unknown package"}</span>
                  <span className="revenue-amount">Ksh. {row.total}</span>
                  <span className="revenue-count">{row.count} payments</span>
                </div>
              ))}
            </section>
            )}

            {adminTab === "subisps" && (
              <section className="dash-card dash-wide">
              <h3>Sub-ISPs</h3>
              <div className="filters">
                <button className={subIspFilter === "all" ? "ghost active" : "ghost"} onClick={() => setSubIspFilter("all")}>
                  All
                </button>
                <button className={subIspFilter === "active" ? "ghost active" : "ghost"} onClick={() => setSubIspFilter("active")}>
                  Active
                </button>
                <button className={subIspFilter === "suspended" ? "ghost active" : "ghost"} onClick={() => setSubIspFilter("suspended")}>
                  Suspended
                </button>
                <button className="pay-btn" onClick={() => openSubReg(null)}>
                  Register Sub-ISP
                </button>
              </div>
              {subIspLoading && <p className="status">Loading Sub-ISPs...</p>}
              {filteredSubIsps.map((sub) => (
                <div key={sub.id} className="revenue-row">
                  <span className="revenue-name">{sub.business || sub.email || "Sub-ISP"}</span>
                  <span className={sub.status === "active" ? "pill active" : "pill inactive"}>{sub.status || "pending"}</span>
                  <button className="ghost" onClick={() => toggleSubIspStatus(sub.id)}>
                    Toggle
                  </button>
                  <select className="select" value={sub.plan || ""} onChange={(e) => updateSubIspPlan(sub.id, e.target.value)}>
                    <option value="">Select plan</option>
                    {subIspPackages.map((pkg) => (
                      <option key={pkg.name} value={pkg.name}>
                        {pkg.name}
                      </option>
                    ))}
                  </select>
                </div>
              ))}
            </section>
            )}

            {adminTab === "users" && (
              <section className="dash-card dash-wide">
              <h3>Users</h3>
              <div className="filters">
                <button className={userFilter === "all" ? "ghost active" : "ghost"} onClick={() => setUserFilter("all")}>
                  All
                </button>
                <button className={userFilter === "active" ? "ghost active" : "ghost"} onClick={() => setUserFilter("active")}>
                  Active
                </button>
                <button className={userFilter === "inactive" ? "ghost active" : "ghost"} onClick={() => setUserFilter("inactive")}>
                  Inactive
                </button>
              </div>
              <div className="table user-table">
                <div className="table-row table-head">
                  <span>Phone</span>
                  <span>Package</span>
                  <span>Status</span>
                  <span>Paid Until</span>
                  <span>Remaining</span>
                  <span>Usage</span>
                </div>
                {users.map((u) => {
                  const usage = usageMap[u.username] || { bytesUsed: 0, timeUsed: 0 };
                  return (
                    <div key={u.phone + "-" + u.email} className="table-row">
                      <span>{u.phone || "-"}</span>
                      <span>{u.package || "-"}</span>
                      <span className={u.active ? "pill active" : "pill inactive"}>{u.active ? "Active" : "Inactive"}</span>
                      <span>{formatRemainingTime(u.paidUntil)}</span>
                      <span>{formatRemaining(u.package, usage.bytesUsed)}</span>
                      <span>
                        {formatBytes(usage.bytesUsed)} / {formatSeconds(usage.timeUsed)}
                        <button className="ghost" onClick={() => fetchUsage(u.username || "")} disabled={usageLoading}>
                          Refresh
                        </button>
                      </span>
                    </div>
                  );
                })}
              </div>
            </section>
            )}
          </div>
        </section>
      ) : (
        <section className="dashboard">
          <header className="dash-header">
            <div>
              <p className="kicker">{isSubIsp ? "Sub-ISP" : "Welcome"}</p>
              <h2 className="dash-title">{isSubIsp ? "Sub-ISP Dashboard" : "Internet Packages"}</h2>
              <p className="subtitle">
                {isSubIsp ? `Account Status: ${subStatusLabel}` : "Choose a package and complete payment via Mpesa prompt."}
              </p>
            </div>
          </header>

          {isSubIsp ? (
            <div className="dash-grid">
              <section className="dash-card compact">
                <h3>Profile</h3>
                {subProfileLoading && <p className="status">Loading profile...</p>}
                {subProfileError && <p className="error">{subProfileError}</p>}
                <p className="muted">{subProfile?.business || "-"}</p>
                <p className="muted">{subProfile?.email || "-"}</p>
                <p className="muted">{subProfile?.phone || "-"}</p>
              </section>
              <section className="dash-card compact">
                <h3>Routers</h3>
                <p className="muted">
                  Plan limit: {subRouterUnlimited ? "Unlimited routers" : `${subRouterCount}/${subRouterLimit} routers used`}
                </p>
                {!subRouterUnlimited && <p className="muted">Remaining slots: {subRouterSlotsLeft}</p>}
                <div className="field">
                  <span>Router Name</span>
                  <input value={routerName} onChange={(e) => setRouterName(e.target.value)} placeholder="Router name" />
                </div>
                <button className="pay-btn" onClick={addRouter} disabled={!canAddSubRouter}>
                  Add Router
                </button>
                {subStatus !== "active" && <p className="error">Router configuration is available only for active plans.</p>}
                {!subRouterUnlimited && subRouterSlotsLeft <= 0 && <p className="error">You have reached your router limit for this plan.</p>}
                {routers.map((router) => (
                  <div key={router.id} className="revenue-row">
                    <span>{router.name}</span>
                    <span>{router.status || "inactive"}</span>
                    <button className="ghost" onClick={() => updateRouterStatus(router.id, router.status === "active" ? "inactive" : "active")}>
                      Toggle
                    </button>
                  </div>
                ))}
              </section>
              <section className="dash-card compact">
                <h3>Usage Lookup</h3>
                <div className="field">
                  <span>Username</span>
                  <input value={subUsageQuery} onChange={(e) => setSubUsageQuery(e.target.value)} placeholder="PPPoE username" />
                </div>
                <button className="pay-btn" onClick={fetchSubUsage} disabled={subUsageLoading}>
                  {subUsageLoading ? "Loading..." : "Check Usage"}
                </button>
                {subUsageError && <p className="error">{subUsageError}</p>}
                {subUsageResult && (
                  <p className="muted">
                    {subUsageResult.username}: {formatBytes(subUsageResult.bytesUsed || 0)} / {formatSeconds(subUsageResult.timeUsedSeconds || 0)}
                  </p>
                )}
              </section>
            </div>
          ) : (
            <>
              <section>
                <h3>Standard Packages</h3>
                <div className="grid">
                  {packages.map((pkg) => (
                    <article key={pkg.duration + "-" + pkg.price + "-" + pkg.tag} className="card" onClick={() => openPay(pkg)}>
                      <p className="duration">{pkg.duration}</p>
                      <p className="price">{pkg.price}</p>
                      <span className="badge">{pkg.tag}</span>
                      <p className="cta">Tap to pay</p>
                    </article>
                  ))}
                </div>
              </section>

              <section className="sub-isp">
                <div className="sub-isp-header">
                  <div>
                    <p className="kicker">Sub-ISP Plans</p>
                    <h3 className="dash-title">Grow Your ISP Business</h3>
                    <p className="subtitle">Pick a plan and register instantly with Mpesa payment.</p>
                  </div>
                </div>
                <div className="grid sub-grid">
                  {subIspPackages.map((pkg) => (
                    <article key={pkg.name} className="card" onClick={() => openSubReg(pkg)}>
                      <p className="duration">{pkg.duration}</p>
                      <p className="price">{pkg.price}</p>
                      <span className="badge">{pkg.name}</span>
                      <p className="muted">
                        {pkg.maxUsers === -1 ? "Unlimited" : pkg.maxUsers} users /{" "}
                        {pkg.maxRouters === -1 ? "Unlimited" : pkg.maxRouters} routers
                      </p>
                      <p className="cta">Tap to register</p>
                    </article>
                  ))}
                </div>
              </section>
            </>
          )}
        </section>
      )}

      {authOpen && (
        <div className="modal-backdrop" role="dialog" aria-modal="true">
          <div className="modal">
            <div className="modal-header">
              <div>
                <p className="modal-kicker">Authentication</p>
                <h2 className="modal-title">Login</h2>
              </div>
              <button className="modal-close" onClick={closeLogin} aria-label="Close">
                x
              </button>
            </div>
            <form className="modal-form" onSubmit={submitLogin}>
              <label className="field">
                <span>Email</span>
                <input type="email" value={authEmail} onChange={(e) => setAuthEmail(e.target.value.trim())} required />
              </label>
              <label className="field">
                <span>Password</span>
                <input type="password" value={authPassword} onChange={(e) => setAuthPassword(e.target.value)} required />
              </label>
              {authError && <p className="error">{authError}</p>}
              <button className="pay-btn" type="submit">
                Login
              </button>
            </form>
          </div>
        </div>
      )}

      {subRegOpen && (
        <div className="modal-backdrop" role="dialog" aria-modal="true">
          <div className="modal">
            <div className="modal-header">
              <div>
                <p className="modal-kicker">Sub-ISP Registration</p>
                <h2 className="modal-title">Register Sub-ISP</h2>
              </div>
              <button className="modal-close" onClick={closeSubReg} aria-label="Close">
                x
              </button>
            </div>

            <form className="modal-form" onSubmit={submitSubReg}>
              <label className="field">
                <span>Business Name</span>
                <input
                  type="text"
                  placeholder="Business name"
                  value={subRegData.business}
                  onChange={(e) => setSubRegData((prev) => ({ ...prev, business: e.target.value }))}
                  required
                />
              </label>
              <label className="field">
                <span>Contact Person</span>
                <input
                  type="text"
                  placeholder="Contact person"
                  value={subRegData.contact}
                  onChange={(e) => setSubRegData((prev) => ({ ...prev, contact: e.target.value }))}
                />
              </label>
              <label className="field">
                <span>Email</span>
                <input
                  type="email"
                  placeholder="you@email.com"
                  value={subRegData.email}
                  onChange={(e) => setSubRegData((prev) => ({ ...prev, email: e.target.value.trim() }))}
                  required
                />
              </label>
              <label className="field">
                <span>Password</span>
                <input
                  type="password"
                  placeholder="At least 8 characters"
                  value={subRegData.password}
                  onChange={(e) => setSubRegData((prev) => ({ ...prev, password: e.target.value }))}
                />
              </label>
              {!subRegData.password && (
                <p className="hint">If you skip a password, your phone number will be used as the default password.</p>
              )}
              {subRegSubmitted && subPasswordError && <p className="error">{subPasswordError}</p>}
              <label className="field">
                <span>Phone</span>
                <input
                  type="tel"
                  placeholder="07XXXXXXXX or 01XXXXXXXX"
                  value={subRegData.phone}
                  onChange={(e) => setSubRegData((prev) => ({ ...prev, phone: e.target.value.trim() }))}
                  required
                />
              </label>
              {subRegSubmitted && subPhoneError && <p className="error">{subPhoneError}</p>}
              <label className="field">
                <span>Location</span>
                <input
                  type="text"
                  placeholder="Town or region"
                  value={subRegData.location}
                  onChange={(e) => setSubRegData((prev) => ({ ...prev, location: e.target.value }))}
                />
              </label>
              <label className="field">
                <span>Package</span>
                <select
                  className="select"
                  value={subRegPackage?.name || ""}
                  onChange={(e) => {
                    const selectedPlan = subIspPackages.find((pkg) => pkg.name === e.target.value) || null;
                    setSubRegPackage(selectedPlan);
                  }}
                  required
                >
                  <option value="" disabled>
                    Select a plan
                  </option>
                  {subIspPackages.map((pkg) => (
                    <option key={pkg.name} value={pkg.name}>
                      {pkg.name} - {pkg.price}
                    </option>
                  ))}
                </select>
              </label>
              {subRegPackage && (
                <p className="hint">
                  {subRegPackage.maxUsers === -1 ? "Unlimited" : subRegPackage.maxUsers} users,{" "}
                  {subRegPackage.maxRouters === -1 ? "Unlimited" : subRegPackage.maxRouters} routers.
                </p>
              )}
              <button className="pay-btn" type="submit" disabled={!canSubmitSubReg || subRegSending}>
                {subRegSending ? "Sending..." : "Register & Pay"}
              </button>
              {subRegStatus && <p className="status">{subRegStatus}</p>}
              <button className="ghost" type="button" onClick={openSubIspLogin}>
                Already registered? Login
              </button>
            </form>
          </div>
        </div>
      )}

      {selected && (
        <div className="modal-backdrop" role="dialog" aria-modal="true">
          <div className="modal">
            <div className="modal-header">
              <div>
                <p className="modal-kicker">Mpesa Payment</p>
                <h2 className="modal-title">Pay {selected.price}</h2>
              </div>
              <button className="modal-close" onClick={closePay} aria-label="Close">
                x
              </button>
            </div>

            <p className="modal-subtitle">
              Package: {selected.duration} - {selected.tag}
            </p>

            <form className="modal-form" onSubmit={submitPay}>
              <label className="field">
                <span>Phone Number</span>
                <input type="tel" placeholder="07XXXXXXXX or 01XXXXXXXX" value={phone} onChange={(e) => setPhone(e.target.value.trim())} />
              </label>
              {submitted && phoneError && <p className="error">{phoneError}</p>}

              <button className="pay-btn" type="submit" disabled={!canPay || sending}>
                {sending ? "Sending..." : "Pay " + selected.price}
              </button>
              <p className="hint">You will receive an Mpesa prompt on your phone.</p>
              {status && <p className="status">{status}</p>}
            </form>
          </div>
        </div>
      )}
    </main>
  );
}
