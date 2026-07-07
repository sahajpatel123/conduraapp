import {Composition} from 'remotion';
import {ConduraConcerto} from './ConduraConcerto';
import {ConduraHypeReel} from './ConduraHypeReel';
import {ConduraMasterpiece, ConduraMasterpieceSchema} from './ConduraMasterpiece';
import './hype.css';
import './masterpiece.css';
import './concerto.css';

export const RemotionRoot = () => {
  return (
    <>
      <Composition
        id="ConduraConcerto"
        component={ConduraConcerto}
        durationInFrames={600}
        fps={30}
        width={1920}
        height={1080}
      />
      <Composition
        id="ConduraHypeReel"
        component={ConduraHypeReel}
        durationInFrames={660}
        fps={30}
        width={1080}
        height={1920}
      />
      <Composition
        id="ConduraMasterpiece"
        component={ConduraMasterpiece}
        durationInFrames={990}
        fps={30}
        width={1920}
        height={1080}
        schema={ConduraMasterpieceSchema}
        defaultProps={{
          brandMark: 'C',
          brandName: 'Condura',
          tagline: 'The conductor for every AI on your computer.',
          problemLines: [
            'Your computer is full of AI.',
            'None of it talks to each other.',
          ],
          chaosStats: ['17 windows', '9 contexts', '0 coordination'],
          conductorClaim: 'Condura conducts every AI on your machine.',
          capabilities: [
            {word: 'MODELS', sub: '12 providers, one key'},
            {word: 'AGENTS', sub: '8 CLIs orchestrated'},
            {word: 'BROWSER', sub: 'clicks, types, reads'},
            {word: 'VOICE', sub: 'say the word'},
            {word: 'LOCAL', sub: 'on your machine first'},
            {word: 'SAFE', sub: 'asks before it acts'},
          ],
          orbitNodes: [
            'Claude',
            'GPT',
            'Gemini',
            'Ollama',
            'Codex',
            'Antigravity',
            'Cursor',
            'OpenCode',
          ],
          cta: 'Free. Forever.',
          site: 'condura.app',
          launch: 'launching soon',
        }}
      />
    </>
  );
};