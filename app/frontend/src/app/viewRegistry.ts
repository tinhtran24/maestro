import {
  BarChart3,
  Bot,
  BrainCircuit,
  CalendarClock,
  FileText,
  FolderTree,
  Kanban,
  Map,
  MessageSquare,
  Settings,
  Workflow,
} from "lucide-react";
import type { NavGroup, ViewId } from "./types";

export const navGroups: NavGroup[] = [
  {
    label: "Workspace",
    items: [
      { id: "chat", label: "Chat", icon: MessageSquare },
      { id: "plan", label: "Plan", icon: FileText },
      { id: "whiteboard", label: "Whiteboard", icon: BrainCircuit },
      { id: "board", label: "Board", icon: Kanban },
      { id: "agent-graph", label: "Agents", icon: Workflow },
      { id: "routines", label: "Routines", icon: CalendarClock },
      { id: "mission", label: "Mission Control", icon: Map },
    ],
  },
  {
    label: "Inspect",
    items: [
      { id: "files", label: "Files", icon: FolderTree },
      { id: "analytics", label: "Analytics", icon: BarChart3 },
      { id: "settings", label: "Settings", icon: Settings },
      { id: "docs", label: "Docs", icon: Bot },
    ],
  },
];

export const navItems = navGroups.flatMap((group) => group.items);

export function isViewId(value: string): value is ViewId {
  return navItems.some((item) => item.id === value);
}
