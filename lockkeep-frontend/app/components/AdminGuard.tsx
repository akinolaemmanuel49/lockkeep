import { Navigate, Outlet } from "react-router";
import { useAuth } from "~/providers/auth";

export default function AdminGuard() {
    const { user, isAuthenticated } = useAuth();

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />;
    }

    if (user?.userType !== "admin") {
        return <Navigate to="/unauthorized" replace />;
    }

    return <Outlet />;
}