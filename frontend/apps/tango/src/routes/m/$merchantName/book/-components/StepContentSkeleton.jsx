export function StepContentSkeleton() {
  return (
    <div className="flex h-full w-full animate-pulse flex-col gap-6">
      <div className="bg-bg-sunken h-9 w-48 rounded"></div>
      <div className="flex flex-1 flex-col">
        <div className="flex flex-1 flex-col gap-4 pt-2">
          {[1, 2, 3, 4].map((i) => (
            <div
              key={i}
              className="bg-bg-raised border-border flex w-full items-start
                justify-between gap-4 rounded-md border p-4"
            >
              <div className="flex w-full items-start gap-4">
                <div className="bg-bg-sunken size-12 shrink-0 rounded-full"></div>
                <div className="flex w-full flex-col gap-3 py-1">
                  <div className="bg-bg-sunken h-5 w-1/3 rounded"></div>
                  <div className="bg-bg-sunken h-4 w-1/4 rounded"></div>
                  <div className="bg-bg-sunken mt-2 h-4 w-16 rounded"></div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
