import { Metadata } from "next";
import StoryboardClient from "./storyboard-client";

export const metadata: Metadata = {
  title: "Component Storyboard | Kloudlite",
  description: "Browse and preview all components in the Kloudlite design system",
};

export default function StoryboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <StoryboardClient>{children}</StoryboardClient>;
}