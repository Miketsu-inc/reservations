/**
 * Free replacement for FullCalendar's premium vertical resource view
 * (https://fullcalendar.io/docs/vertical-resource-view), built on the
 * protected API that FullCalendar v7 exposes for its own premium plugins.
 *
 * Supported:
 * - views: resourceTimeGrid, resourceTimeGridDay, resourceTimeGridWeek
 * - calendar options: resources, resourceOrder, datesAboveResources,
 *   eventResourceEditable, resourceDayHeader{Content,Class,InnerClass,Align,DidMount,WillUnmount}
 * - resource props: id, title, businessHours, eventColor, eventContrastColor,
 *   eventClass, eventConstraint, eventOverlap, eventAllow, extendedProps
 * - event props: resourceId, resourceIds, resourceEditable
 * - `resource` on dateClick/select info, `oldResource`/`newResource` on eventDrop
 * - `dayMinWidth`, which otherwise needs the premium scrollgrid plugin
 *
 * Without resources it renders a plain time grid with every event merged, so
 * switching between team and plain grids keeps the same mounted view.
 */
import timeGridPlugin from "@fullcalendar/react/timegrid";
import {
  AllDaySplitter,
  buildDateDataConfigs,
  buildDateRenderConfig,
  buildDayColsFromSeries,
  compareByFieldSpecs,
  createDayHeaderFormatter,
  createEventUi,
  DateComponent,
  DaySeriesModel,
  DaySeriesSlicer,
  DayTimeColsSlicer,
  identity,
  isPropsValid,
  memoize,
  mergeEventStores,
  NowTimer,
  organizeSegsByCol,
  parseBusinessHours,
  parseFieldSpecs,
  refineClassNameGenerator,
  splitInteractionByCol,
  Splitter,
  TimeGridLayout,
} from "@fullcalendar/react/protected-api";

const EMPTY_SEGS = [];
const EMPTY_RESOURCES = { list: [], byId: new Map() };
const EMPTY_SLICED = {
  allDay: { bgEventSegs: EMPTY_SEGS },
  timed: { bgEventSegs: EMPTY_SEGS },
};

// Resources

const RESOURCE_PROPS = new Set([
  "id",
  "title",
  "businessHours",
  "eventColor",
  "eventContrastColor",
  "eventClass",
  "eventConstraint",
  "eventOverlap",
  "eventAllow",
  "extendedProps",
]);

function parseResource(input, context) {
  const id = input.id == null ? "" : String(input.id);
  const title = input.title == null ? "" : String(input.title);
  const extendedProps = { ...input.extendedProps };

  for (const propName in input) {
    if (!RESOURCE_PROPS.has(propName)) {
      extendedProps[propName] = input[propName];
    }
  }

  return {
    id,
    title,
    businessHours: input.businessHours
      ? parseBusinessHours(input.businessHours, context)
      : null,
    ui: createEventUi(
      {
        color: input.eventColor,
        contrastColor: input.eventContrastColor,
        class: input.eventClass,
        constraint: input.eventConstraint,
        overlap: input.eventOverlap,
        allow: input.eventAllow,
      },
      context
    ),
    api: Object.freeze({ id, title, extendedProps }),
  };
}

function sortResources(resources, resourceOrder) {
  const fieldSpecs = parseFieldSpecs(resourceOrder ?? "id");

  // Sort on the public shape so resourceOrder can reference extendedProps.
  return resources
    .map((resource) => ({
      resource,
      fields: { ...resource.api.extendedProps, ...resource.api },
    }))
    .sort((a, b) => compareByFieldSpecs(a.fields, b.fields, fieldSpecs))
    .map(({ resource }) => resource);
}

// Plugin hooks only receive the calendar context, so parsed resources are
// cached per `resources` array instead of being stored in calendar state.
const resourcesCache = new WeakMap();

function getResources(context) {
  const { options, dateEnv } = context;
  const input = options.resources;
  if (!Array.isArray(input)) return EMPTY_RESOURCES;

  const cached = resourcesCache.get(input);
  if (
    cached &&
    cached.dateEnv === dateEnv &&
    cached.resourceOrder === options.resourceOrder
  ) {
    return cached.resources;
  }

  const list = sortResources(
    input.map((resourceInput) => parseResource(resourceInput, context)),
    options.resourceOrder
  );
  const resources = {
    list,
    byId: new Map(list.map((resource) => [resource.id, resource])),
  };

  resourcesCache.set(input, {
    dateEnv,
    resourceOrder: options.resourceOrder,
    resources,
  });
  return resources;
}

function getResourceApi(context, resourceId) {
  return getResources(context).byId.get(resourceId)?.api ?? null;
}

// Events

function parseEventResourceIds(refined) {
  const resourceIds = [...(refined.resourceIds ?? []), refined.resourceId];

  return [
    ...new Set(
      resourceIds
        .filter((resourceId) => resourceId != null)
        .map((resourceId) => String(resourceId))
    ),
  ];
}

function isResourceEditable(eventDef, context) {
  return (
    eventDef.resourceEditable ?? context.options.eventResourceEditable ?? true
  );
}

function getAppliedResourceMutation(mutation, eventDef, context) {
  const resourceMutation = mutation.resourceMutation;

  if (
    !resourceMutation ||
    !isResourceEditable(eventDef, context) ||
    !eventDef.resourceIds.includes(resourceMutation.matchResourceId)
  ) {
    return null;
  }

  return resourceMutation;
}

/**
 * Splits view props into one bucket per resource. Events and selections
 * without a resource land in the "" bucket, as does everything when there are
 * no resources.
 */
class ResourceSplitter extends Splitter {
  constructor(resources) {
    super();
    this.resources = resources;
  }

  getKeyInfo() {
    const keyInfo = { "": {} };
    for (const resource of this.resources) {
      keyInfo[resource.id] = {
        businessHours: resource.businessHours,
        ui: resource.ui,
      };
    }
    return keyInfo;
  }

  getKeysForDateSpan(dateSpan) {
    return [dateSpan.resourceId ?? ""];
  }

  getKeysForEventDef(eventDef) {
    return this.resources.length && eventDef.resourceIds?.length
      ? eventDef.resourceIds
      : [""];
  }
}

function isResourcePropsValid(state, context) {
  const splitProps = new ResourceSplitter(
    getResources(context).list
  ).splitProps(state);
  const unassignedProps = splitProps[""];

  for (const resourceId in splitProps) {
    let props = splitProps[resourceId];

    // Events without a resource block every resource.
    if (resourceId && unassignedProps) {
      props = {
        ...props,
        eventStore: mergeEventStores(
          unassignedProps.eventStore,
          props.eventStore
        ),
        eventUiBases: {
          ...unassignedProps.eventUiBases,
          ...props.eventUiBases,
        },
      };
    }

    const dateSpanMeta = resourceId ? { resourceId } : {};
    if (!isPropsValid(props, context, dateSpanMeta)) {
      return false;
    }
  }

  return true;
}

// Column layout

function buildColMapper(resourceCnt, dayCnt, datesAboveResources) {
  return datesAboveResources
    ? (resourceIndex, dayIndex) => dayIndex * resourceCnt + resourceIndex
    : (resourceIndex, dayIndex) => resourceIndex * dayCnt + dayIndex;
}

function remapTimedSeg(seg, resourceIndex, colMapper) {
  return [{ ...seg, col: colMapper(resourceIndex, seg.col) }];
}

/**
 * All-day segs span a range of day indexes. When dates are above resources
 * those days are no longer adjacent, so the seg is cut wherever the mapped
 * columns stop being contiguous.
 */
function remapAllDaySeg(seg, resourceIndex, colMapper) {
  const pieces = [];
  let piece = null;

  for (let dayIndex = seg.start; dayIndex < seg.end; dayIndex++) {
    const col = colMapper(resourceIndex, dayIndex);

    if (piece && piece.end === col) {
      piece.end = col + 1;
    } else {
      piece = {
        ...seg,
        start: col,
        end: col + 1,
        isStart: seg.isStart && dayIndex === seg.start,
      };
      pieces.push(piece);
    }
    piece.isEnd = seg.isEnd && dayIndex === seg.end - 1;
  }

  return pieces;
}

function joinSegs(remapSeg, colMapper, sharedSegs, ...segsByResource) {
  const joined = [];

  segsByResource.forEach((segs, resourceIndex) => {
    for (const seg of sharedSegs.concat(segs)) {
      joined.push(...remapSeg(seg, resourceIndex, colMapper));
    }
  });

  return joined;
}

function joinInteractions(remapSeg, colMapper, ...interactionsByResource) {
  let joined = null;

  interactionsByResource.forEach((interaction, resourceIndex) => {
    if (!interaction) return;

    joined ??= {
      affectedInstances: {},
      segs: [],
      isEvent: interaction.isEvent,
    };
    Object.assign(joined.affectedInstances, interaction.affectedInstances);
    for (const seg of interaction.segs) {
      joined.segs.push(...remapSeg(seg, resourceIndex, colMapper));
    }
  });

  return joined;
}

/**
 * Merges per-resource slicer output into a single set of column-indexed segs.
 * Every field is memoized separately so unchanged resources keep the
 * downstream TimeGridLayout memoization intact.
 */
class ResourceSegJoiner {
  constructor(remapSeg) {
    this.remapSeg = remapSeg;
    this.joinFgEventSegs = memoize(joinSegs);
    this.joinBgEventSegs = memoize(joinSegs);
    this.joinBusinessHourSegs = memoize(joinSegs);
    this.joinDateSelectionSegs = memoize(joinSegs);
    this.joinEventDrag = memoize(joinInteractions);
    this.joinEventResize = memoize(joinInteractions);
  }

  join(colMapper, slicedByResource, unassignedSliced) {
    const { remapSeg } = this;
    const pluck = (field) => slicedByResource.map((sliced) => sliced[field]);

    return {
      fgEventSegs: this.joinFgEventSegs(
        remapSeg,
        colMapper,
        EMPTY_SEGS,
        ...pluck("fgEventSegs")
      ),
      // Background events without a resource render in every resource.
      bgEventSegs: this.joinBgEventSegs(
        remapSeg,
        colMapper,
        unassignedSliced.bgEventSegs,
        ...pluck("bgEventSegs")
      ),
      businessHourSegs: this.joinBusinessHourSegs(
        remapSeg,
        colMapper,
        EMPTY_SEGS,
        ...pluck("businessHourSegs")
      ),
      dateSelectionSegs: this.joinDateSelectionSegs(
        remapSeg,
        colMapper,
        EMPTY_SEGS,
        ...pluck("dateSelectionSegs")
      ),
      eventDrag: this.joinEventDrag(remapSeg, colMapper, ...pluck("eventDrag")),
      eventResize: this.joinEventResize(
        remapSeg,
        colMapper,
        ...pluck("eventResize")
      ),
    };
  }
}

function buildCells(dayCols, resources, datesAboveResources, colMapper) {
  if (!resources.length) return dayCols;

  const cells = [];

  resources.forEach((resource, resourceIndex) => {
    dayCols.forEach((dayCol, dayIndex) => {
      cells[colMapper(resourceIndex, dayIndex)] = {
        ...dayCol,
        key: `${resource.id}:${dayCol.key}`,
        isMajor: datesAboveResources ? resourceIndex === 0 : dayIndex === 0,
        renderProps: { ...dayCol.renderProps, resource: resource.api },
        attrs: { ...dayCol.attrs, "data-resource-id": resource.id },
        dateSpanProps: { ...dayCol.dateSpanProps, resourceId: resource.id },
      };
    });
  });

  return cells;
}

// Header

function buildResourceRenderConfig(context) {
  const { options } = context;

  return {
    generatorName: "resourceDayHeaderContent",
    customGenerator: options.resourceDayHeaderContent,
    classNameGenerator: options.resourceDayHeaderClass,
    innerClassNameGenerator: options.resourceDayHeaderInnerClass,
    didMount: options.resourceDayHeaderDidMount,
    willUnmount: options.resourceDayHeaderWillUnmount,
    align: options.resourceDayHeaderAlign,
    sticky: options._dayHeaderSticky,
  };
}

function buildResourceDataConfig(
  resource,
  { colSpan = 1, isMajor, keyPrefix = "", date, context }
) {
  return {
    key: keyPrefix + resource.id,
    renderProps: {
      resource: resource.api,
      text: resource.title,
      date,
      isMajor,
      isDisabled: false,
      inPopover: false,
      view: context.viewApi,
    },
    attrs: { "data-resource-id": resource.id },
    colSpan,
  };
}

function buildHeaderTiers(
  resources,
  dayDates,
  datesAboveResources,
  dateProfile,
  todayRange,
  dayHeaderFormat,
  context
) {
  const dayCnt = dayDates.length;
  const resourceRenderConfig = buildResourceRenderConfig(context);
  const dateRenderConfig = buildDateRenderConfig(
    dayHeaderFormat,
    true,
    context
  );

  if (!resources.length) {
    return [
      {
        isDateRow: true,
        renderConfig: dateRenderConfig,
        dataConfigs: buildDateDataConfigs(
          dayDates,
          true,
          dateProfile,
          todayRange,
          dayHeaderFormat,
          context
        ),
      },
    ];
  }

  if (datesAboveResources) {
    return [
      {
        isDateRow: true,
        renderConfig: dateRenderConfig,
        dataConfigs: buildDateDataConfigs(
          dayDates,
          true,
          dateProfile,
          todayRange,
          dayHeaderFormat,
          context,
          resources.length,
          "",
          {},
          {},
          "",
          1
        ),
      },
      {
        isDateRow: false,
        renderConfig: resourceRenderConfig,
        dataConfigs: dayDates.flatMap((dayDate, dayIndex) =>
          resources.map((resource, resourceIndex) =>
            buildResourceDataConfig(resource, {
              isMajor: resourceIndex === 0,
              keyPrefix: `${dayIndex}:`,
              date: context.dateEnv.toDate(dayDate),
              context,
            })
          )
        ),
      },
    ];
  }

  return [
    {
      isDateRow: false,
      renderConfig: resourceRenderConfig,
      dataConfigs: resources.map((resource) =>
        buildResourceDataConfig(resource, {
          colSpan: dayCnt,
          isMajor: true,
          context,
        })
      ),
    },
    {
      isDateRow: true,
      renderConfig: dateRenderConfig,
      dataConfigs: resources.flatMap((resource) =>
        buildDateDataConfigs(
          dayDates,
          true,
          dateProfile,
          todayRange,
          dayHeaderFormat,
          context,
          1,
          `${resource.id}:`,
          { resource: resource.api },
          { "data-resource-id": resource.id },
          "",
          dayCnt,
          dayCnt
        )
      ),
    },
  ];
}

// Scrolling

/**
 * Keeps sibling scrollers (header, all-day row, body, footer scrollbar) at
 * the same offset along one axis. FullCalendar requires this for
 * `dayMinWidth` but only ships an implementation in its premium scrollgrid
 * plugin.
 */
class ScrollerSyncer {
  constructor(isHorizontal = false) {
    this.isHorizontal = isHorizontal;
    this.bindings = [];
    // Only the scroller the user is moving drives the others; the scroll
    // events triggered by syncing the followers are ignored.
    this.leader = null;
    this.startListeners = new Set();
    this.scrollListeners = new Set();
    this.endListeners = new Set();
  }

  get x() {
    return (this.leader ?? this.bindings[0]?.scroller)?.x ?? 0;
  }

  get y() {
    return (this.leader ?? this.bindings[0]?.scroller)?.y ?? 0;
  }

  scrollTo(coords) {
    for (const { scroller } of this.bindings) {
      scroller.scrollTo(coords);
    }
  }

  handleChildren(scrollers) {
    const mounted = scrollers.filter((scroller) => scroller?.listener);
    const isUnchanged =
      mounted.length === this.bindings.length &&
      mounted.every(
        (scroller, i) =>
          this.bindings[i].scroller === scroller &&
          this.bindings[i].listener === scroller.listener
      );
    if (isUnchanged) return;

    this.unbind();
    this.bindings = mounted.map((scroller) => this.bind(scroller));
  }

  bind(scroller) {
    const { listener } = scroller;
    const handlers = {
      scrollStart: (isDevice) => {
        if (this.leader) return;

        this.leader = scroller;
        for (const handler of this.startListeners) handler(isDevice);
      },
      scroll: (isDevice) => {
        if (this.leader !== scroller) return;

        const pos = this.isHorizontal ? scroller.x : scroller.y;
        const coords = this.isHorizontal ? { x: pos } : { y: pos };
        for (const binding of this.bindings) {
          if (binding.scroller !== scroller) binding.scroller.scrollTo(coords);
        }
        for (const handler of this.scrollListeners) handler(isDevice, pos);
      },
      scrollEnd: (isDevice) => {
        if (this.leader !== scroller) return;

        this.leader = null;
        for (const handler of this.endListeners) handler(isDevice);
      },
    };

    for (const type in handlers) {
      listener.emitter.on(type, handlers[type]);
    }
    return { scroller, listener, handlers };
  }

  unbind() {
    for (const { listener, handlers } of this.bindings) {
      for (const type in handlers) {
        listener.emitter.off(type, handlers[type]);
      }
    }
    this.bindings = [];
    this.leader = null;
  }

  destroy() {
    this.unbind();
  }

  addScrollStartListener(handler) {
    this.startListeners.add(handler);
  }

  removeScrollStartListener(handler) {
    this.startListeners.delete(handler);
  }

  addScrollListener(handler) {
    this.scrollListeners.add(handler);
  }

  removeScrollListener(handler) {
    this.scrollListeners.delete(handler);
  }

  addScrollEndListener(handler) {
    this.endListeners.add(handler);
  }

  removeScrollEndListener(handler) {
    this.endListeners.delete(handler);
  }
}

// View

function isSameResourceHit(hit0, hit1) {
  return hit0.dateSpan.resourceId === hit1.dateSpan.resourceId;
}

function createResourceSlicers() {
  return {
    allDaySplitter: new AllDaySplitter(),
    daySeriesSlicer: new DaySeriesSlicer(),
    dayTimeColsSlicer: new DayTimeColsSlicer(),
  };
}

class ResourceTimeGridView extends DateComponent {
  constructor(...args) {
    super(...args);

    this.buildResourceSplitter = memoize(
      (resources) => new ResourceSplitter(resources)
    );
    this.buildDaySeries = memoize(
      (dateProfile, dateProfileGenerator) =>
        new DaySeriesModel(dateProfile.renderRange, dateProfileGenerator)
    );
    this.buildDayCols = memoize(buildDayColsFromSeries);
    this.extractColDates = memoize((cols) => cols.map((col) => col.date));
    this.extractColRanges = memoize((cols) => cols.map((col) => col.range));
    this.buildColMapper = memoize(buildColMapper);
    this.buildCells = memoize(buildCells);
    this.createDayHeaderFormatter = memoize(createDayHeaderFormatter);
    this.buildHeaderTiers = memoize(buildHeaderTiers);

    this.allDayJoiner = new ResourceSegJoiner(remapAllDaySeg);
    this.timedJoiner = new ResourceSegJoiner(remapTimedSeg);

    this.splitFgEventSegs = memoize(organizeSegsByCol);
    this.splitBgEventSegs = memoize(organizeSegsByCol);
    this.splitBusinessHourSegs = memoize(organizeSegsByCol);
    this.splitNowIndicatorSegs = memoize(organizeSegsByCol);
    this.splitDateSelectionSegs = memoize(organizeSegsByCol);
    this.splitEventDrag = memoize(splitInteractionByCol);
    this.splitEventResize = memoize(splitInteractionByCol);

    // Slicers memoize on their last input, so each resource needs its own.
    this.slicersByResourceId = new Map();
    this.nowSlicer = new DayTimeColsSlicer();
  }

  getSlicers(resourceId) {
    let slicers = this.slicersByResourceId.get(resourceId);
    if (!slicers) {
      slicers = createResourceSlicers();
      this.slicersByResourceId.set(resourceId, slicers);
    }
    return slicers;
  }

  sliceResourceProps(resourceId, props, daySeries, dayRanges) {
    const { options } = this.context;
    const { dateProfile } = this.props;
    const slicers = this.getSlicers(resourceId);
    const splitProps = slicers.allDaySplitter.splitProps(props);

    return {
      allDay: slicers.daySeriesSlicer.sliceProps(
        splitProps.allDay,
        dateProfile,
        options.nextDayThreshold,
        this.context,
        daySeries
      ),
      timed: slicers.dayTimeColsSlicer.sliceProps(
        splitProps.timed,
        dateProfile,
        null,
        this.context,
        dayRanges
      ),
    };
  }

  render() {
    const { props, context } = this;
    const { dateProfile } = props;
    const { options, dateProfileGenerator, dateEnv } = context;
    const datesAboveResources = Boolean(options.datesAboveResources);

    const resources = getResources(context).list;
    const daySeries = this.buildDaySeries(dateProfile, dateProfileGenerator);
    const dayCols = this.buildDayCols(daySeries, dateEnv, {
      slotRange: dateProfile,
      activeRange: dateProfile.activeRange,
    });
    const dayDates = this.extractColDates(dayCols);
    const dayRanges = this.extractColRanges(dayCols);
    const colMapper = this.buildColMapper(
      Math.max(resources.length, 1),
      dayCols.length,
      datesAboveResources
    );
    const cells = this.buildCells(
      dayCols,
      resources,
      datesAboveResources,
      colMapper
    );

    const splitProps = this.buildResourceSplitter(resources).splitProps(props);
    const unassignedSliced = this.sliceResourceProps(
      "",
      splitProps[""],
      daySeries,
      dayRanges
    );
    // Without resources the unassigned bucket is the only column group.
    const slicedByResource = resources.length
      ? resources.map((resource) =>
          this.sliceResourceProps(
            resource.id,
            splitProps[resource.id],
            daySeries,
            dayRanges
          )
        )
      : [unassignedSliced];
    // Unassigned background events repeat in every resource, which would
    // duplicate them when that bucket is already the only column group.
    const unassignedBg = resources.length ? unassignedSliced : EMPTY_SLICED;

    const allDayProps = this.allDayJoiner.join(
      colMapper,
      slicedByResource.map((sliced) => sliced.allDay),
      unassignedBg.allDay
    );
    const timedProps = this.timedJoiner.join(
      colMapper,
      slicedByResource.map((sliced) => sliced.timed),
      unassignedBg.timed
    );

    const dayHeaderFormat = this.createDayHeaderFormatter(
      options.dayHeaderFormat,
      true,
      dayCols.length
    );

    return (
      <NowTimer unit={options.nowIndicator ? "minute" : "day"}>
        {(nowDate, todayRange, nowMs) => {
          const colCount = cells.length;
          const nowIndicatorSegs =
            !props.forPrint && options.nowIndicator
              ? joinSegs(
                  remapTimedSeg,
                  colMapper,
                  this.nowSlicer.sliceNowDate(
                    nowDate,
                    dateProfile,
                    options.nextDayThreshold,
                    context,
                    dayRanges
                  ),
                  ...slicedByResource.map(() => EMPTY_SEGS)
                )
              : null;

          return (
            <TimeGridLayout
              labelId={props.labelId}
              labelStr={props.labelStr}
              dateProfile={dateProfile}
              nowDate={nowDate}
              nowMs={nowMs}
              todayRange={todayRange}
              cells={cells}
              forPrint={props.forPrint}
              className={props.className}
              isHitComboAllowed={isSameResourceHit}
              headerTiers={this.buildHeaderTiers(
                resources,
                dayDates,
                datesAboveResources,
                dateProfile,
                todayRange,
                dayHeaderFormat,
                context
              )}
              fgEventSegs={allDayProps.fgEventSegs}
              bgEventSegs={allDayProps.bgEventSegs}
              businessHourSegs={allDayProps.businessHourSegs}
              dateSelectionSegs={allDayProps.dateSelectionSegs}
              eventDrag={allDayProps.eventDrag}
              eventResize={allDayProps.eventResize}
              fgEventSegsByCol={this.splitFgEventSegs(
                timedProps.fgEventSegs,
                colCount
              )}
              bgEventSegsByCol={this.splitBgEventSegs(
                timedProps.bgEventSegs,
                colCount
              )}
              businessHourSegsByCol={this.splitBusinessHourSegs(
                timedProps.businessHourSegs,
                colCount
              )}
              nowIndicatorSegsByCol={this.splitNowIndicatorSegs(
                nowIndicatorSegs,
                colCount
              )}
              dateSelectionSegsByCol={this.splitDateSelectionSegs(
                timedProps.dateSelectionSegs,
                colCount
              )}
              eventDragByCol={this.splitEventDrag(
                timedProps.eventDrag,
                colCount
              )}
              eventResizeByCol={this.splitEventResize(
                timedProps.eventResize,
                colCount
              )}
              eventSelection={props.eventSelection}
            />
          );
        }}
      </NowTimer>
    );
  }
}

// Plugin

function buildResourceSpanProps(dateSpan, context) {
  return dateSpan.resourceId == null
    ? {}
    : { resource: getResourceApi(context, dateSpan.resourceId) };
}

const resourceTimeGridPlugin = {
  name: "resource-timegrid",
  deps: [timeGridPlugin],
  optionRefiners: {
    resources: identity,
    resourceOrder: identity,
    datesAboveResources: Boolean,
    eventResourceEditable: Boolean,
    resourceDayHeaderContent: identity,
    resourceDayHeaderClass: refineClassNameGenerator,
    resourceDayHeaderInnerClass: refineClassNameGenerator,
    resourceDayHeaderAlign: identity,
    resourceDayHeaderDidMount: identity,
    resourceDayHeaderWillUnmount: identity,
  },
  eventRefiners: {
    // Stringified by parseEventResourceIds, which also drops null ids.
    resourceId: identity,
    resourceIds: identity,
    resourceEditable: Boolean,
  },
  eventDefMemberAdders: [
    (refined) => ({
      resourceIds: parseEventResourceIds(refined),
      resourceEditable: refined.resourceEditable,
    }),
  ],
  eventDragMutationMassagers: [
    (mutation, hit0, hit1) => {
      const matchResourceId = hit0.dateSpan.resourceId;
      const setResourceId = hit1.dateSpan.resourceId;

      if (
        matchResourceId &&
        setResourceId &&
        matchResourceId !== setResourceId
      ) {
        mutation.resourceMutation = { matchResourceId, setResourceId };
      }
    },
  ],
  eventDefMutationAppliers: [
    (eventDef, mutation, context) => {
      const resourceMutation = getAppliedResourceMutation(
        mutation,
        eventDef,
        context
      );
      if (!resourceMutation) return;

      const { matchResourceId, setResourceId } = resourceMutation;
      eventDef.resourceIds = [
        ...new Set(
          eventDef.resourceIds.map((resourceId) =>
            resourceId === matchResourceId ? setResourceId : resourceId
          )
        ),
      ];
    },
  ],
  eventDropTransformers: [
    (mutation, context, eventDef) => {
      const resourceMutation = getAppliedResourceMutation(
        mutation,
        eventDef,
        context
      );
      if (!resourceMutation) return {};

      return {
        oldResource: getResourceApi(context, resourceMutation.matchResourceId),
        newResource: getResourceApi(context, resourceMutation.setResourceId),
      };
    },
  ],
  dateSelectionTransformers: [
    (hit0, hit1) => {
      const resourceId0 = hit0.dateSpan.resourceId;
      const resourceId1 = hit1.dateSpan.resourceId;

      if (resourceId0 == null && resourceId1 == null) return undefined;
      if (resourceId0 !== resourceId1) return false;
      return { resourceId: resourceId0 };
    },
  ],
  datePointTransforms: [buildResourceSpanProps],
  dateSpanTransforms: [buildResourceSpanProps],
  isPropsValid: isResourcePropsValid,
  scrollerSyncerClass: ScrollerSyncer,
  views: {
    resourceTimeGrid: {
      type: "timeGrid",
      component: ResourceTimeGridView,
    },
    resourceTimeGridDay: {
      type: "resourceTimeGrid",
      duration: { days: 1 },
    },
    resourceTimeGridWeek: {
      type: "resourceTimeGrid",
      duration: { weeks: 1 },
    },
  },
};

export default resourceTimeGridPlugin;
