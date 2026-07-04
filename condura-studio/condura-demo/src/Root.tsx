import {Composition} from 'remotion';
import {ConduraCinematicLaunch} from './ConduraCinematicLaunch';
import {ConduraDemo} from './ConduraDemo';
import {ConduraMinimalLaunch} from './ConduraMinimalLaunch';
import {ConduraPrelaunch} from './ConduraPrelaunch';
import './style.css';
import './prelaunch.css';
import './minimal.css';
import './cinematic.css';

export const RemotionRoot = () => {
  return (
    <>
      <Composition
        id="ConduraDemo"
        component={ConduraDemo}
        durationInFrames={1800}
        fps={30}
        width={1920}
        height={1080}
      />
      <Composition
        id="ConduraPrelaunch"
        component={ConduraPrelaunch}
        durationInFrames={1620}
        fps={30}
        width={1920}
        height={1080}
      />
      <Composition
        id="ConduraMinimalLaunch"
        component={ConduraMinimalLaunch}
        durationInFrames={1680}
        fps={30}
        width={1920}
        height={1080}
      />
      <Composition
        id="ConduraCinematicLaunch"
        component={ConduraCinematicLaunch}
        durationInFrames={810}
        fps={30}
        width={1920}
        height={1440}
      />
    </>
  );
};
