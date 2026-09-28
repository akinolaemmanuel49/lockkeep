import VaultDial from "./VaultDial";
import Nameplate from "./Nameplate";

export default function AuthPlate({
  title,
  eyebrow,
  description,
  children,
}: {
  title: string;
  eyebrow: string;
  description: string;
  children: React.ReactNode;
}) {
  return (
    <div className="relative w-full max-w-md">
      <VaultDial
        size={420}
        className="pointer-events-none absolute -right-28 -top-28 opacity-[0.13]"
      />
      <div className="plate relative p-8">
        <div className="mb-6 text-center">
          <Nameplate className="justify-center">{eyebrow}</Nameplate>
          <h1 className="mt-3 font-display text-2xl font-semibold tracking-tight text-ivory">
            {title}
          </h1>
          <p className="mt-1 text-sm text-sand-500">{description}</p>
        </div>
        {children}
      </div>
    </div>
  );
}