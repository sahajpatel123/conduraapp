import React from "react";
import { Composition } from "remotion";
import { DemoVideo } from "./DemoVideo";
import { TOTAL, FPS_VALUE } from "./constants";

export const RemotionRoot: React.FC = () => {
  return (
    <Composition
      id="Demo"
      component={DemoVideo}
      durationInFrames={TOTAL}
      fps={FPS_VALUE}
      width={1920}
      height={1080}
    />
  );
};
