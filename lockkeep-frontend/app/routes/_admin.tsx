import { Outlet, redirect } from "react-router";
import type { Route } from "./+types/_admin";

export default function AdminLayout() {
    return <Outlet />;
}

// Client-side guard: redirect if not admin
export function clientLoader({ request }: Route.ClientLoaderArgs) {
    // This runs in the browser after hydration
    const userStr = sessionStorage.getItem("vault_user");
    if (!userStr) {
        throw redirect("/login");
    }

    const user = JSON.parse(userStr);
    if (user.userType !== "admin") {
        throw redirect("/unauthorized");
    }

    return { user };
}