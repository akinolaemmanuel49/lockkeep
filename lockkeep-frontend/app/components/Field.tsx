export default function Field({
  label,
  type,
  value,
  onChange,
  required,
  placeholder,
}: {
  label: string;
  type: string;
  value: string;
  onChange: (val: string) => void;
  required?: boolean;
  placeholder?: string;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-sm font-medium text-slate-400">{label}</label>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        required={required}
        placeholder={placeholder}
        className="rounded-lg border border-slate-700 bg-slate-950 px-4 py-2.5 text-sm text-slate-200 placeholder:text-slate-600 focus:border-sky-400 focus:outline-none"
      />
    </div>
  );
}
