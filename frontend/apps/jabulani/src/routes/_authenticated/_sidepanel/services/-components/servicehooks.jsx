import { useCallback } from "react";

export function normalizeServicePhases(phases) {
  return [...phases]
    .sort((a, b) => (a.sequence || 0) - (b.sequence || 0))
    .map((phase, index) => ({
      ...phase,
      sequence: index + 1,
    }));
}

export function useServicePhases(setServiceData) {
  const addPhase = useCallback(
    (newPhase) => {
      setServiceData((prev) => {
        const phases = normalizeServicePhases(prev.phases);

        return {
          ...prev,
          phases: [
            ...phases,
            {
              ...newPhase,
              id: -1,
              sequence: phases.length + 1,
            },
          ],
        };
      });
    },
    [setServiceData]
  );

  const updatePhase = useCallback(
    (updatedPhase) => {
      setServiceData((prev) => ({
        ...prev,
        phases: prev.phases.map((phase) =>
          phase.sequence === updatedPhase.sequence
            ? { ...phase, ...updatedPhase }
            : phase
        ),
      }));
    },
    [setServiceData]
  );

  const removePhase = useCallback(
    (sequence) => {
      setServiceData((prev) => {
        const phases = prev.phases.filter(
          (phase) => phase.sequence !== sequence
        );

        return {
          ...prev,
          phases: normalizeServicePhases(phases),
        };
      });
    },
    [setServiceData]
  );

  const reorderPhase = useCallback(
    (sequence, direction) => {
      setServiceData((prev) => {
        const phases = [...prev.phases].sort(
          (a, b) => (a.sequence || 0) - (b.sequence || 0)
        );
        const currentIndex = phases.findIndex(
          (phase) => phase.sequence === sequence
        );
        const targetIndex = currentIndex + direction;

        if (
          currentIndex === -1 ||
          targetIndex < 0 ||
          targetIndex >= phases.length
        ) {
          return prev;
        }

        [phases[currentIndex], phases[targetIndex]] = [
          phases[targetIndex],
          phases[currentIndex],
        ];

        return {
          ...prev,
          phases: phases.map((phase, index) => ({
            ...phase,
            sequence: index + 1,
          })),
        };
      });
    },
    [setServiceData]
  );

  return {
    addPhase,
    updatePhase,
    removePhase,
    reorderPhase,
  };
}
