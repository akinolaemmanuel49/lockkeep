export const strengthLabels = [
    "Very Weak",
    "Weak",
    "Fair",
    "Good",
    "Strong",
    "Very Strong",
];

export const strengthColors = [
    "bg-red-500",
    "bg-red-400",
    "bg-amber-400",
    "bg-lime-400",
    "bg-green-400",
    "bg-emerald-500",
];

export const calculatePasswordStrength = (pw: string): number => {
    let score = 0;
    if (pw.length >= 12) score += 1;
    if (pw.length >= 16) score += 1;
    if (/[A-Z]/.test(pw)) score += 1;
    if (/[0-9]/.test(pw)) score += 1;
    if (/[^A-Za-z0-9]/.test(pw)) score += 1;
    return score;
};