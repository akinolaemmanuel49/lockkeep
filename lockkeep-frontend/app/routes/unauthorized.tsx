import { Link } from "react-router";

export default function Unauthorized() {
    return (
        <div className="flex flex-col items-center justify-center py-20">
            <h1 className="mb-4 text-3xl font-bold text-slate-100">403</h1>
            <p className="mb-6 text-slate-400">You don't have permission to access this page.</p>
            <Link
                to="/dashboard"
                className="rounded-lg bg-sky-400 px-4 py-2 font-medium text-slate-950 hover:bg-sky-300"
            >
                Back to Dashboard
            </Link>
        </div>
    );
}