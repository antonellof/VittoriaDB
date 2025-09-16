import { cn } from "@/lib/utils";
// import type { Experimental_GeneratedImage } from "ai";

export type ImageProps = {
  base64?: string;
  url?: string;
  uint8Array?: Uint8Array;
  mediaType?: string;
  className?: string;
  alt?: string;
};

export const Image = ({
  base64,
  uint8Array,
  mediaType,
  ...props
}: ImageProps) => (
  <img
    {...props}
    alt={props.alt}
    className={cn(
      "h-auto max-w-full overflow-hidden rounded-md",
      props.className
    )}
    src={`data:${mediaType};base64,${base64}`}
  />
);
