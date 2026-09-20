import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";

const ScrollSpyContext = createContext(null);
const ActiveSectionContext = createContext(null);

function findScrollParent(element) {
  let parent = element.parentElement;

  while (parent) {
    const { overflow, overflowY } = getComputedStyle(parent);
    if (/(auto|scroll|overlay)/.test(overflowY || overflow)) {
      return parent;
    }
    parent = parent.parentElement;
  }

  return window;
}

function getSectionOrder(sections) {
  return [...sections.entries()]
    .sort(([, first], [, second]) => {
      if (first.element === second.element) return 0;

      return first.element.compareDocumentPosition(second.element) &
        Node.DOCUMENT_POSITION_FOLLOWING
        ? -1
        : 1;
    })
    .map(([id]) => id);
}

function getScrollBounds(scrollTarget) {
  if (scrollTarget === window) {
    return { bottom: window.innerHeight, top: 0 };
  }

  const top = scrollTarget.getBoundingClientRect().top;
  return { bottom: top + scrollTarget.clientHeight, top };
}

function isAtScrollEnd(scrollTarget) {
  if (scrollTarget === window) {
    return (
      window.scrollY + window.innerHeight >=
      Math.max(
        document.documentElement.scrollHeight,
        document.body.scrollHeight
      ) -
        1
    );
  }

  return (
    scrollTarget.scrollTop + scrollTarget.clientHeight >=
    scrollTarget.scrollHeight - 1
  );
}

function getActiveSectionId({
  currentId,
  scrollOffset,
  scrollTarget,
  sectionIds,
  sections,
}) {
  const { bottom, top } = getScrollBounds(scrollTarget);

  if (isAtScrollEnd(scrollTarget)) {
    const visibleIds = sectionIds.filter((id) => {
      const section = sections.get(id);
      if (!section) return false;

      const rect = section.element.getBoundingClientRect();
      return rect.bottom > top && rect.top < bottom;
    });

    return visibleIds.includes(currentId)
      ? currentId
      : (visibleIds[0] ?? sectionIds[sectionIds.length - 1]);
  }

  const readingLine = top + scrollOffset;
  let nextId = sectionIds[0];

  for (const id of sectionIds) {
    const section = sections.get(id);
    if (!section) continue;

    if (section.element.getBoundingClientRect().top <= readingLine) {
      nextId = id;
    } else {
      break;
    }
  }

  return nextId;
}

export function ScrollSpyProvider({ children, scrollOffset = 64 }) {
  const sectionsRef = useRef(new Map());
  const activeIdRef = useRef(null);
  const [sectionIds, setSectionIds] = useState([]);
  const [activeId, setActiveId] = useState(null);
  const offset = Math.max(0, Number(scrollOffset) || 0);

  const setActiveSection = useCallback((id) => {
    activeIdRef.current = id;
    setActiveId((currentId) => (currentId === id ? currentId : id));
  }, []);

  const registerSection = useCallback(
    (id, label, element) => {
      const previous = sectionsRef.current.get(id);
      const changed =
        !previous || previous.label !== label || previous.element !== element;

      sectionsRef.current.set(id, { element, label });
      if (!changed) return;

      const nextIds = getSectionOrder(sectionsRef.current);
      setSectionIds(nextIds);
      setActiveSection(
        activeIdRef.current && sectionsRef.current.has(activeIdRef.current)
          ? activeIdRef.current
          : (nextIds[0] ?? null)
      );
    },
    [setActiveSection]
  );

  const unregisterSection = useCallback(
    (id) => {
      if (!sectionsRef.current.delete(id)) return;

      const nextIds = getSectionOrder(sectionsRef.current);
      setSectionIds(nextIds);
      setActiveSection(
        activeIdRef.current === id ? (nextIds[0] ?? null) : activeIdRef.current
      );
    },
    [setActiveSection]
  );

  const getLabel = useCallback(
    (id) => sectionsRef.current.get(id)?.label ?? id,
    []
  );

  useEffect(() => {
    if (sectionIds.length === 0) return undefined;

    const firstSection = sectionsRef.current.get(sectionIds[0]);
    if (!firstSection) return undefined;

    const scrollTarget = findScrollParent(firstSection.element);
    let frameId = null;

    function updateActiveSection() {
      if (frameId !== null) return;

      frameId = requestAnimationFrame(() => {
        frameId = null;
        setActiveSection(
          getActiveSectionId({
            currentId: activeIdRef.current,
            scrollOffset: offset,
            scrollTarget,
            sectionIds,
            sections: sectionsRef.current,
          })
        );
      });
    }

    scrollTarget.addEventListener("scroll", updateActiveSection, {
      passive: true,
    });
    window.addEventListener("resize", updateActiveSection);
    updateActiveSection();

    return () => {
      scrollTarget.removeEventListener("scroll", updateActiveSection);
      window.removeEventListener("resize", updateActiveSection);
      if (frameId !== null) cancelAnimationFrame(frameId);
    };
  }, [offset, sectionIds, setActiveSection]);

  const contextValue = useMemo(
    () => ({
      getLabel,
      registerSection,
      scrollOffset: offset,
      sectionIds,
      selectSection: setActiveSection,
      unregisterSection,
    }),
    [
      getLabel,
      registerSection,
      offset,
      sectionIds,
      setActiveSection,
      unregisterSection,
    ]
  );

  return (
    <ScrollSpyContext.Provider value={contextValue}>
      <ActiveSectionContext.Provider value={activeId}>
        {children}
      </ActiveSectionContext.Provider>
    </ScrollSpyContext.Provider>
  );
}

function useScrollSpy() {
  const context = useContext(ScrollSpyContext);
  if (!context) {
    throw new Error(
      "ScrollSpy components must be used within a <ScrollSpyProvider>"
    );
  }

  return context;
}

function useScrollSpyActiveId() {
  return useContext(ActiveSectionContext);
}

export function ScrollSpySection({
  id,
  label,
  styles,
  style,
  children,
  ...props
}) {
  const { registerSection, scrollOffset, unregisterSection } = useScrollSpy();
  const ref = useRef(null);

  useEffect(() => {
    const element = ref.current;
    if (!element) return undefined;

    registerSection(id, label, element);
    return () => unregisterSection(id);
  }, [id, label, registerSection, unregisterSection]);

  return (
    <div
      id={id}
      ref={ref}
      className={styles}
      style={{ ...style, scrollMarginTop: scrollOffset + "px" }}
      {...props}
    >
      {children}
    </div>
  );
}

export function ScrollSpyNav() {
  const { getLabel, sectionIds, selectSection } = useScrollSpy();
  const activeId = useScrollSpyActiveId();

  if (sectionIds.length === 0) return null;

  return (
    <div
      className="border-border_color bg-bg_color sticky top-0 z-10 w-full
        min-w-0 border-b py-2.5 md:self-start md:overflow-y-auto md:border-0
        md:py-0"
    >
      <nav
        aria-label="Section navigation"
        className="md:border-border_color md:bg-layer_bg flex w-full min-w-0
          scrollbar-thin flex-row gap-1.5 overflow-x-auto px-4 md:flex-col
          md:gap-1 md:overflow-visible md:rounded-lg md:border md:p-2"
      >
        {sectionIds.map((id) => {
          const active = id === activeId;

          return (
            <a
              key={id}
              href={"#" + encodeURIComponent(id)}
              onClick={() => selectSection(id)}
              aria-current={active ? "location" : undefined}
              className={
                active
                  ? `bg-primary/10 text-primary shrink-0 rounded-lg px-3 py-2
                    text-sm whitespace-nowrap md:w-full md:py-2.5`
                  : `text-text_color/70 hover:bg-primary/10 shrink-0 rounded-lg
                    px-3 py-2 text-sm whitespace-nowrap md:w-full md:py-2.5`
              }
            >
              {getLabel(id)}
            </a>
          );
        })}
      </nav>
    </div>
  );
}
