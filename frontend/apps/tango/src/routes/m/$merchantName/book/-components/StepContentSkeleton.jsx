export function StepContentSkeleton() {
  return (
    <div className="flex h-full w-full animate-pulse flex-col gap-6">
      <div className="h-9 w-48 rounded bg-gray-200 dark:bg-gray-700"></div>
      <div className="flex flex-1 flex-col">
        <div className="flex flex-1 flex-col gap-4 pt-2">
          {[1, 2, 3, 4].map((i) => (
            <div
              key={i}
              className="bg-layer_bg border-border_color flex w-full items-start
                justify-between gap-4 rounded-md border p-4"
            >
              <div className="flex w-full items-start gap-4">
                <div
                  className="size-12 shrink-0 rounded-full bg-gray-200
                    dark:bg-gray-700"
                ></div>
                <div className="flex w-full flex-col gap-3 py-1">
                  <div
                    className="h-5 w-1/3 rounded bg-gray-200 dark:bg-gray-700"
                  ></div>
                  <div
                    className="h-4 w-1/4 rounded bg-gray-100 dark:bg-gray-800"
                  ></div>
                  <div
                    className="mt-2 h-4 w-16 rounded bg-gray-200
                      dark:bg-gray-700"
                  ></div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
