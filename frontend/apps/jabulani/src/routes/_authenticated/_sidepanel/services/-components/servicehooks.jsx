import { useCallback } from "react";

export function useServicePhases(setServiceData) {
  const addPhase = useCallback(
    (newPhase) => {
      setServiceData((prev) => {
        const maxSequence =
          prev.phases.length > 0
            ? Math.max(...prev.phases.map((p) => p.sequence || 0))
            : 0;

        return {
          ...prev,
          phases: [
            ...prev.phases,
            {
              ...newPhase,
              id: 0,
              sequence: maxSequence + 1,
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
        const filteredPhases = prev.phases.filter(
          (phase) => phase.sequence !== sequence
        );

        const reSequencedPhases = filteredPhases.map((phase, index) => ({
          ...phase,
          sequence: index + 1,
        }));

        return {
          ...prev,
          phases: reSequencedPhases,
        };
      });
    },
    [setServiceData]
  );

  return {
    addPhase,
    updatePhase,
    removePhase,
  };
}
