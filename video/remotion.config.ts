import { Config } from "@remotion/cli/config";

/*
  Synaptic demo — render configuration.
  Targets the brief's delivery spec: MP4 / H.264 / 1080p / 30 fps,
  visually-lossless CRF, yuv420p for broad compatibility.
*/
Config.setVideoImageFormat("jpeg");
Config.setCodec("h264");
Config.setCrf(18);
Config.setPixelFormat("yuv420p");
Config.setOverwriteOutput(true);
Config.setConcurrency(null); // let Remotion choose based on cores
