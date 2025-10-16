"use client";

import { cn } from "@/lib/utils";
import { type ComponentProps, memo } from "react";
import { Streamdown } from "streamdown";

type ResponseProps = ComponentProps<typeof Streamdown>;

export const Response = memo(
  ({ className, ...props }: ResponseProps) => (
    <Streamdown
      className={cn(
        "response-content w-full max-w-full overflow-hidden [&>*:first-child]:mt-0 [&>*:last-child]:mb-0",
        // Prevent horizontal overflow for all content types
        "break-words [&>pre]:overflow-x-auto [&>code]:break-all",
        // Handle HTML content specifically
        "[&>div]:max-w-full [&>div]:overflow-hidden",
        // Handle tables and wide content
        "[&>table]:w-full [&>table]:table-fixed [&>table_td]:break-words",
        // Handle images and media
        "[&>img]:max-w-full [&>img]:h-auto",
        className
      )}
      {...props}
    />
  ),
  (prevProps, nextProps) => prevProps.children === nextProps.children
);

Response.displayName = "Response";
