export default function AccordionSection({
  id,
  title,
  isOpen,
  onToggle,
  children,
}: {
  id: string;
  title: string;
  isOpen: boolean;
  onToggle: () => void;
  children: React.ReactNode;
}) {
  return (
    <div className="mb-4 rounded-2xl border border-slate-800 bg-slate-900 overflow-hidden">
      <button
        onClick={onToggle}
        className="flex w-full items-center justify-between px-6 py-4 text-left hover:bg-slate-800/50 transition-colors"
        aria-expanded={isOpen}
        aria-controls={`section-${id}`}
      >
        <h2 className="text-lg font-semibold text-slate-100">{title}</h2>
        <svg
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className={`text-slate-500 transition-transform duration-200 ${isOpen ? "rotate-180" : ""}`}
        >
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </button>
      <div
        id={`section-${id}`}
        className={`overflow-hidden transition-all duration-200 ${isOpen ? "max-h-[2000px] opacity-100" : "max-h-0 opacity-0"}`}
      >
        <div className="px-6 pb-6 mt-4">{children}</div>
      </div>
    </div>
  );
}
