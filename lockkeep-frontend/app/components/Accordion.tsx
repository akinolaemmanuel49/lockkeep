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
    <div className="plate mb-4 overflow-hidden">
      <button
        onClick={onToggle}
        className="flex w-full items-center justify-between px-6 py-4 text-left transition-colors hover:bg-brass-500/5"
        aria-expanded={isOpen}
        aria-controls={`section-${id}`}
      >
        <h2 className="font-display text-lg font-medium text-ivory">{title}</h2>
        <span className="text-sand-500 transition-transform duration-200">
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.8"
            strokeLinecap="round"
            strokeLinejoin="round"
            className={isOpen ? "rotate-180" : ""}
          >
            <polyline points="6 9 12 15 18 9" />
          </svg>
        </span>
      </button>
      <div
        id={`section-${id}`}
        className={`overflow-hidden transition-all duration-200 ${
          isOpen ? "max-h-[2000px] opacity-100" : "max-h-0 opacity-0"
        }`}
      >
        <div className="border-t border-brass-500/10 px-6 pb-6 pt-5">
          {children}
        </div>
      </div>
    </div>
  );
}