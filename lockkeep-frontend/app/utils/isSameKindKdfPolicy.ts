import type { KDFParams } from "~/types";

export function isSameKdfPolicy(
  current: Omit<KDFParams, "salt">,
  user: KDFParams,
): boolean {
  return (
    current.algorithm === user.algorithm &&
    current.iterations === user.iterations &&
    current.memory === user.memory &&
    current.parallelism === user.parallelism
  );
}