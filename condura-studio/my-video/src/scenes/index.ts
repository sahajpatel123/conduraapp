// src/scenes/index.ts
//
// Central registry: every scene exports its component *and* a meta object
// matching `SceneSpec` in `video.config.ts`. Video.tsx imports only this
// file — never the individual scene folders — so we can refactor freely.

import { ColdOpen, coldOpenMeta } from "./ColdOpen";
import {
  SystemEmergence,
  systemEmergenceMeta,
} from "./SystemEmergence";
import { Conductor, conductorMeta } from "./Conductor";
import { Sovereignty, sovereigntyMeta } from "./Sovereignty";
import { Constellation, constellationMeta } from "./Constellation";
import { Close, closeMeta } from "./Close";

export const sceneRegistry = [
  { Component: ColdOpen, meta: coldOpenMeta },
  { Component: SystemEmergence, meta: systemEmergenceMeta },
  { Component: Conductor, meta: conductorMeta },
  { Component: Sovereignty, meta: sovereigntyMeta },
  { Component: Constellation, meta: constellationMeta },
  { Component: Close, meta: closeMeta },
] as const;

export type RegisteredScene = (typeof sceneRegistry)[number];
