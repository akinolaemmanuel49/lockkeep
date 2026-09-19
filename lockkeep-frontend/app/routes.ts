import { type RouteConfig, index, layout, route } from "@react-router/dev/routes";

export default [
    index("routes/home.tsx"),
    route("login", "routes/login.tsx"),
    route("signup", "routes/signup.tsx"),
    route("setup", "routes/setup.tsx"),
    route("dashboard", "routes/dashboard.tsx"),
    route("settings", "routes/settings.tsx"),
    route("callback", "routes/callback.tsx"),
    layout("routes/_admin.tsx", [
        route("admin/policy", "routes/_admin-policy.tsx"),
    ]),
    route("*", "routes/not-found.tsx"),
    route("unauthorized", "routes/unauthorized.tsx"),
] satisfies RouteConfig;
