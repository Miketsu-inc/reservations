import BackArrowIcon from "../assets/icons/BackArrowIcon";

export default function SidePanelProfile({ image, text, closeSidePanel }) {
  return (
    <div className="flex flex-row items-center">
      <img className="basis-1/8 rounded-full" src={image} />
      <span className="ml-2 basis-auto">{text}</span>
      <span
        className="basis-1/8 ml-auto rounded-md hover:bg-hvr_gray sm:hidden"
        onClick={closeSidePanel}
      >
        <BackArrowIcon styles="h-6 w-6" />
      </span>
    </div>
  );
}
