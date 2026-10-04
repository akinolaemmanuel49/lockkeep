export const strengthLabels = [
    "Very Weak",
    "Weak",
    "Fair",
    "Good",
    "Strong",
    "Very Strong",
];

export const strengthColors = [
    "bg-vermillion-500",
    "bg-vermillion-400",
    "bg-vermillion-400",
    "bg-brass-400",
    "bg-patina-400",
    "bg-patina-500",
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