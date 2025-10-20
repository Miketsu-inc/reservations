export default function SettingsHeader() {
  return (
    <header className="mb-6 mt-1 w-full">
      <div className="flex w-full flex-row">
        <div className="w-16">
          <img
            className="h-auto w-full rounded-3xl object-cover"
            src="https://dummyimage.com/200x200/d156c3/000000.jpg"
          />
        </div>
        <div className="flex flex-col justify-center pl-5 lg:gap-2">
          <h1 className="text-xl font-bold lg:text-4xl">
            {/* {merchantInfo.merchant_name} */}
            Bwnet
          </h1>
          <p className="text-sm">Your personal account</p>
        </div>
      </div>
    </header>
  );
}
