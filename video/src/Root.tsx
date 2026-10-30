import React from "react";
import {
  AbsoluteFill,
  Audio,
  Composition,
  Easing,
  Img,
  Sequence,
  interpolate,
  spring,
  staticFile,
  useCurrentFrame,
  useVideoConfig,
} from "remotion";

const FPS = 30;
const BG = "#0d1417";
const PANEL = "#16211f";
const INK = "#eef4f1";
const MUTED = "#9fb3ab";
const BRAND = "#0e7c6b";
const BRAND_L = "#2ea88f";
const S1 = "#e4b53a";
const S2 = "#e8714a";
const FONT =
  "-apple-system, 'Segoe UI', Roboto, 'Be Vietnam Pro', system-ui, sans-serif";
const MONO = "ui-monospace, Menlo, Consolas, monospace";

// scene durations in frames (VO seconds + padding, measured from vo files)
export const SCENES_EN = [264, 565, 736, 562, 313, 656, 810, 308];
export const SCENES_VI = [286, 589, 842, 594, 310, 678, 810, 337];
const totalFramesEN = SCENES_EN.reduce((a, b) => a + b, 0);
const totalFramesVI = SCENES_VI.reduce((a, b) => a + b, 0);

const fade = (frame: number, from: number, len = 18) =>
  interpolate(frame, [from, from + len], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

const Title: React.FC<{ children: React.ReactNode; size?: number }> = ({
  children,
  size = 84,
}) => (
  <div
    style={{
      fontFamily: FONT,
      fontSize: size,
      fontWeight: 800,
      color: INK,
      lineHeight: 1.12,
      letterSpacing: -1,
    }}
  >
    {children}
  </div>
);

const Body: React.FC<{ children: React.ReactNode; size?: number }> = ({
  children,
  size = 34,
}) => (
  <div
    style={{
      fontFamily: FONT,
      fontSize: size,
      color: MUTED,
      lineHeight: 1.5,
      maxWidth: 1300,
    }}
  >
    {children}
  </div>
);

const Scene1: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const s = spring({ frame, fps, config: { damping: 18 } });
  return (
    <AbsoluteFill
      style={{
        backgroundColor: BG,
        justifyContent: "center",
        padding: "0 140px",
      }}
    >
      <div style={{ opacity: s, transform: `translateY(${30 - 30 * s}px)` }}>
        <div
          style={{
            fontFamily: MONO,
            color: BRAND_L,
            fontSize: 28,
            marginBottom: 28,
          }}
        >
          AGENTSPILLCHECK
        </div>
        <Title>Your AI agent reads everything.</Title>
        <div style={{ marginTop: 34, opacity: fade(frame, 70) }}>
          <Body size={40}>What if it leaks what it finds?</Body>
        </div>
      </div>
    </AbsoluteFill>
  );
};

const ProblemCard: React.FC<{
  title: string;
  text: string;
  delay: number;
  accent: string;
}> = ({ title, text, delay, accent }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const s = spring({ frame: frame - delay, fps, config: { damping: 20 } });
  return (
    <div
      style={{
        flex: 1,
        backgroundColor: PANEL,
        borderRadius: 24,
        padding: "42px 38px",
        opacity: s,
        transform: `translateY(${26 - 26 * s}px)`,
      }}
    >
      <div
        style={{
          width: 64,
          height: 10,
          borderRadius: 5,
          backgroundColor: accent,
          marginBottom: 30,
        }}
      />
      <div
        style={{
          fontFamily: FONT,
          fontSize: 40,
          fontWeight: 700,
          color: INK,
          marginBottom: 16,
        }}
      >
        {title}
      </div>
      <Body size={28}>{text}</Body>
    </div>
  );
};

const Scene2: React.FC = () => (
  <AbsoluteFill style={{ backgroundColor: BG, padding: "110px 120px" }}>
    <Title size={64}>Agents are helpful. Sometimes too helpful.</Title>
    <div style={{ display: "flex", gap: 28, marginTop: 90 }}>
      <ProblemCard
        delay={26}
        accent={S1}
        title="Careless summary"
        text="It prints an API key into its own answer."
      />
      <ProblemCard
        delay={52}
        accent={S2}
        title="Helpful forward"
        text="It mails a contract to the wrong place."
      />
      <ProblemCard
        delay={78}
        accent="#9b82d6"
        title="Hidden instruction"
        text="A web page tells it to send everything it knows."
      />
    </div>
  </AbsoluteFill>
);

const FlowStep: React.FC<{
  label: string;
  text: string;
  delay: number;
}> = ({ label, text, delay }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const s = spring({ frame: frame - delay, fps, config: { damping: 20 } });
  return (
    <div
      style={{
        flex: 1,
        opacity: s,
        transform: `scale(${0.92 + 0.08 * s})`,
        backgroundColor: PANEL,
        borderRadius: 22,
        padding: "36px 30px",
        borderTop: `8px solid ${BRAND}`,
      }}
    >
      <div
        style={{
          fontFamily: FONT,
          fontSize: 38,
          fontWeight: 800,
          color: INK,
          marginBottom: 12,
        }}
      >
        {label}
      </div>
      <Body size={26}>{text}</Body>
    </div>
  );
};

const Scene3: React.FC = () => (
  <AbsoluteFill style={{ backgroundColor: BG, padding: "110px 120px" }}>
    <Title size={64}>Test the agent where nothing is real.</Title>
    <div style={{ display: "flex", gap: 24, marginTop: 90 }}>
      <FlowStep delay={20} label="1. Baits" text="Fake secrets planted in files and mail." />
      <FlowStep delay={50} label="2. Fake stage" text="Drive, mail, blog and shop that look real." />
      <FlowStep delay={80} label="3. Detectors" text="Every read, send and save is checked." />
      <FlowStep delay={110} label="4. Report" text="A grade, a timeline, and fixes." />
    </div>
  </AbsoluteFill>
);

const Scene4: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ backgroundColor: BG, padding: "110px 120px" }}>
      <Title size={64}>Deploy in seconds.</Title>
      <div
        style={{
          marginTop: 70,
          backgroundColor: "#0a0f0e",
          borderRadius: 22,
          padding: "42px 46px",
          fontFamily: MONO,
          fontSize: 32,
          color: "#cfe6de",
          lineHeight: 1.75,
          opacity: fade(frame, 20),
        }}
      >
        <div>$ spillcheck demo --profile leaky</div>
        <div>$ docker compose -f deploy/docker-compose.yml up -d</div>
        <div>$ spillcheck mcp config --client claude-desktop</div>
        <div style={{ color: BRAND_L }}>
          spill_start → spill_next_task → stage_* → spill_finish
        </div>
      </div>
    </AbsoluteFill>
  );
};

const RUN_LINES = [
  "$ spillcheck demo --profile leaky --locale en",
  "session: sess-42",
  "grade: B",
  "score: 86",
  "findings: 3",
];

const Scene5: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill
      style={{
        backgroundColor: BG,
        justifyContent: "center",
        padding: "0 200px",
      }}
    >
      <div
        style={{
          backgroundColor: "#0a0f0e",
          borderRadius: 22,
          padding: "54px 58px",
          fontFamily: MONO,
          fontSize: 44,
          color: "#cfe6de",
          lineHeight: 1.9,
          minHeight: 480,
        }}
      >
        {RUN_LINES.map((line, i) => {
          const start = 14 + i * 38;
          const shown = Math.max(
            0,
            Math.min(line.length, Math.floor((frame - start) / 1.6))
          );
          if (frame < start) return null;
          const isCmd = i === 0;
          const color =
            i === 2 ? S1 : i === 4 ? S2 : isCmd ? "#ffffff" : "#cfe6de";
          return (
            <div key={i} style={{ color, fontWeight: i >= 2 ? 700 : 400 }}>
              {line.slice(0, shown)}
              {shown < line.length ? "▌" : ""}
            </div>
          );
        })}
      </div>
    </AbsoluteFill>
  );
};

const Ticks: React.FC<{ kept: number; lost: { sev: string; n: number }[] }> = ({
  kept,
  lost,
}) => {
  const frame = useCurrentFrame();
  const bars: string[] = [];
  for (let i = 0; i < kept; i++) bars.push("#e6efec");
  for (const l of lost) for (let i = 0; i < l.n; i++) bars.push(l.sev);
  return (
    <div style={{ display: "flex", gap: 3, height: 54, alignItems: "flex-end" }}>
      {bars.map((c, i) => {
        const grown = interpolate(frame, [10 + i * 0.35, 26 + i * 0.35], [0, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
        });
        return (
          <div
            key={i}
            style={{
              flex: 1,
              height: `${(i + 1) % 10 === 0 ? 64 : 54}px`,
              backgroundColor: c,
              borderRadius: 3,
              transform: `scaleY(${grown})`,
              transformOrigin: "bottom",
            }}
          />
        );
      })}
    </div>
  );
};

const Scene6: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const s = spring({ frame, fps, config: { damping: 20 } });
  return (
    <AbsoluteFill style={{ backgroundColor: BG, padding: "96px 120px" }}>
      <div style={{ display: "flex", gap: 40, opacity: s }}>
        <div
          style={{
            width: 190,
            height: 190,
            borderRadius: 24,
            backgroundColor: "#1d3a35",
            color: "#3fb39a",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            fontFamily: FONT,
            fontSize: 110,
            fontWeight: 800,
          }}
        >
          B
        </div>
        <div style={{ paddingTop: 22 }}>
          <div style={{ fontFamily: FONT, fontSize: 64, fontWeight: 800, color: INK }}>
            86 <span style={{ color: MUTED, fontSize: 34 }}>/ 100</span>
          </div>
          <Body size={30}>86 points kept, 3 baits escaped</Body>
        </div>
      </div>
      <div style={{ marginTop: 70 }}>
        <Ticks kept={86} lost={[{ sev: S1, n: 6 }, { sev: S2, n: 8 }]} />
      </div>
      <div style={{ display: "flex", gap: 22, marginTop: 64 }}>
        {[
          { sev: "S1 ×2", c: S1, t: "printed into the answer" },
          { sev: "S2 ×1", c: S2, t: "sent by email" },
          { sev: "S3 ×0", c: "#d2546a", t: "unknown destination" },
        ].map((x, i) => (
          <div
            key={x.sev}
            style={{
              flex: 1,
              backgroundColor: PANEL,
              borderRadius: 20,
              padding: "30px 32px",
              opacity: fade(frame, 60 + i * 26),
            }}
          >
            <div style={{ fontFamily: FONT, fontSize: 36, fontWeight: 800, color: x.c }}>
              {x.sev}
            </div>
            <Body size={26}>{x.t}</Body>
          </div>
        ))}
      </div>
    </AbsoluteFill>
  );
};

// Report tour: full-page screenshot of the real EN report.html,
// panned top to bottom once, one narration beat per section.
// Image is 1280x3450 shown at 1920 wide (x1.5), so content is 5175 tall.
const TourScene: React.FC<{ lang?: string }> = ({ lang = "en" }) => {
  const frame = useCurrentFrame();
  const MAXSHIFT = 5175 - 1080;
  const stops: [number, number][] = [
    [0, 0], [60, 0], [150, 0.18], [210, 0.18], [330, 0.45], [390, 0.45],
    [510, 0.72], [570, 0.72], [690, 1], [810, 1],
  ];
  const frac = interpolate(
    frame,
    stops.map(([f]) => f),
    stops.map(([, v]) => v),
    { easing: Easing.inOut(Easing.ease), extrapolateLeft: "clamp", extrapolateRight: "clamp" }
  );
  const pre = lang === "vi" ? "vo-vitour" : "vo-tour";
  const beats: [number, string][] = [
    [10, `${pre}1.wav`], [160, `${pre}2.wav`], [340, `${pre}3.wav`],
    [520, `${pre}4.wav`], [670, `${pre}5.wav`],
  ];
  return (
    <AbsoluteFill style={{ backgroundColor: "#f6f7f5", overflow: "hidden" }}>
      <Img
        src={staticFile("report-pan.png")}
        style={{
          width: 1920,
          height: 5175,
          transform: `translateY(${-frac * MAXSHIFT}px)`,
        }}
      />
      {beats.map(([from, src]) => (
        <Sequence key={src} from={from} durationInFrames={150}>
          <Audio src={staticFile(src)} />
        </Sequence>
      ))}
    </AbsoluteFill>
  );
};

const Scene7: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const s = spring({ frame, fps, config: { damping: 16 } });
  return (
    <AbsoluteFill
      style={{
        backgroundColor: BG,
        justifyContent: "center",
        alignItems: "center",
        textAlign: "center",
        padding: "0 160px",
      }}
    >
      <div style={{ opacity: s, transform: `scale(${0.94 + 0.06 * s})` }}>
        <div
          style={{
            width: 84,
            height: 84,
            backgroundColor: BRAND,
            borderRadius: 20,
            margin: "0 auto 36px",
          }}
        />
        <Title size={72}>AgentSpillCheck</Title>
        <div style={{ marginTop: 26 }}>
          <Body size={36}>
            Test your agent's discipline before a real leak tests it for you.
          </Body>
        </div>
        <div
          style={{
            marginTop: 30,
            fontFamily: MONO,
            color: BRAND_L,
            fontSize: 24,
            opacity: fade(frame, 60),
          }}
        >
          NO REAL DATA IS EVER TOUCHED
        </div>
      </div>
    </AbsoluteFill>
  );
};

// caption chunks per scene, weighted by narration timing
export const CAPTIONS: { scene: number; chunks: string[] }[] = [
  {
    scene: 1,
    chunks: [
      "Your AI agent reads your files, your email, and the web.",
      "But what if it leaks what it finds?",
    ],
  },
  {
    scene: 2,
    chunks: [
      "Agents summarize folders, forward mail, and follow instructions hidden in web pages.",
      "One careless summary can print an API key.",
      "One helpful forward can mail a contract away.",
      "And a hidden instruction can make the agent send everything it knows to a stranger.",
    ],
  },
  {
    scene: 3,
    chunks: [
      "AgentSpillCheck tests that behavior safely.",
      "It plants fake secrets, called baits, in a fake workspace.",
      "The agent works there like a normal job, reading files and sending mail.",
      "Detectors watch every read, every send, and every file the agent writes.",
      "At the end, you get a graded report: what was read, where it leaked, and why.",
    ],
  },
  {
    scene: 4,
    chunks: [
      "Deployment takes seconds.",
      "Install the binary, or start the Docker image.",
      "Connect your agent over MCP with one config snippet.",
      "It receives tasks, works with fake drive, mail, blog, and shop tools, and finishes like a real assignment.",
    ],
  },
  {
    scene: 5,
    chunks: [
      "Here is a real run.",
      "A scripted careless agent reads the environment file, prints the key, and mails the contract onward.",
    ],
  },
  {
    scene: 6,
    chunks: [
      "The report grades the session.",
      "Eighty-six of one hundred points were kept.",
      "Three baits escaped: two printed into the answer, one sent by email.",
      "Every finding shows the masked value, the timeline, the root cause, and how to fix it.",
      "A clean agent scores one hundred and earns grade A.",
    ],
  },
  {
    scene: 7,
    chunks: [
      "Here is the full report, from the top.",
      "The severity spread shows how dangerous each leak was.",
      "Every finding lists the timeline and the fix.",
      "Scoring and coverage show exactly where the points went.",
      "And one clear recommendation: what to fix first.",
    ],
  },
  {
    scene: 8,
    chunks: [
      "AgentSpillCheck.",
      "Test your agent's discipline before a real leak tests it for you.",
      "No real data is ever touched.",
    ],
  },
];

const CAPTION_SPANS: Record<number, [number, number][]> = {
  1: [[0, 132], [132, 264]],
  2: [[0, 190], [190, 320], [320, 435], [435, 565]],
  3: [[0, 120], [120, 300], [300, 460], [460, 600], [600, 736]],
  4: [[0, 90], [90, 260], [260, 380], [380, 562]],
  5: [[0, 110], [110, 313]],
  6: [[0, 100], [100, 230], [230, 360], [360, 520], [520, 656]],
  7: [[0, 160], [160, 340], [340, 520], [520, 670], [670, 810]],
  8: [[0, 70], [70, 200], [200, 308]],
};

const Captions: React.FC<{ scene: number }> = ({ scene }) => {
  const frame = useCurrentFrame();
  const chunks = CAPTIONS[scene - 1].chunks;
  const spans = CAPTION_SPANS[scene];
  const idx = spans.findIndex(([a, b]) => frame >= a && frame < b);
  if (idx < 0) return null;
  return (
    <div
      style={{
        position: "absolute",
        left: 0,
        right: 0,
        bottom: 64,
        display: "flex",
        justifyContent: "center",
        pointerEvents: "none",
      }}
    >
      <div
        style={{
          backgroundColor: "rgba(5,10,9,0.82)",
          color: "#ffffff",
          fontFamily: FONT,
          fontSize: 32,
          lineHeight: 1.35,
          padding: "16px 30px",
          borderRadius: 14,
          maxWidth: 1400,
          textAlign: "center",
        }}
      >
        {chunks[idx]}
      </div>
    </div>
  );
};

export const DemoVideo: React.FC<{ lang?: string }> = ({ lang = "en" }) => {
  const vi = lang === "vi";
  const scenes = vi ? SCENES_VI : SCENES_EN;
  const audioOf = (i: number): string | null => {
    if (i === 6) return null;
    const n = i < 6 ? i + 1 : 7;
    return vi ? `vo-vi${n}.wav` : `vo${n}.wav`;
  };
  const sceneNodes = [
    <Scene1 key="s1" />, <Scene2 key="s2" />, <Scene3 key="s3" />,
    <Scene4 key="s4" />, <Scene5 key="s5" />, <Scene6 key="s6" />,
    <TourScene key="tour" lang={lang} />, <Scene7 key="s7" />,
  ];
  let from = 0;
  return (
    <AbsoluteFill style={{ backgroundColor: BG }}>
      {sceneNodes.map((node, i) => {
        const dur = scenes[i];
        const audio = audioOf(i);
        const el = (
          <Sequence key={i} from={from} durationInFrames={dur}>
            {node}
            {audio ? <Audio src={staticFile(audio)} /> : null}
            {vi ? null : <Captions scene={i + 1} />}
          </Sequence>
        );
        from += dur;
        return el;
      })}
    </AbsoluteFill>
  );
};

export const RemotionRoot: React.FC = () => (
  <>
    <Composition
      id="DemoVideo"
      component={DemoVideo}
      durationInFrames={totalFramesEN}
      fps={FPS}
      width={1920}
      height={1080}
      defaultProps={{ lang: "en" }}
    />
    <Composition
      id="DemoVideoVI"
      component={DemoVideo}
      durationInFrames={totalFramesVI}
      fps={FPS}
      width={1920}
      height={1080}
      defaultProps={{ lang: "vi" }}
    />
  </>
);
