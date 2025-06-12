export default function ServiceCategoryItem({ children, category }) {
  return (
    <>
      {category.services.length > 0 && category.id !== null ? (
        <div className="py-6">
          <div className="bg-secondary flex flex-row items-center gap-4 rounded-lg p-2 shadow-md">
            <div className="flex h-14 shrink-0 overflow-hidden rounded-lg md:h-24">
              <img
                className="size-full object-cover"
                src="https://dummyimage.com/160x100/d156c3/000000.jpg"
                alt="service category photo"
              />
            </div>
            <p className="text-lg font-semibold">{`${category.id ? `${category.name}` : "Uncategorized"}`}</p>
          </div>
          <div className="px-4">{children}</div>
        </div>
      ) : (
        children
      )}
    </>
  );
}
