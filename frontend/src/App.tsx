import './index.css';

function App() {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      <header className="border-b border-slate-800 bg-slate-900/60 backdrop-blur">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
          <div className="text-lg font-bold tracking-tight">Auto-ToS</div>
          <span className="text-sm text-slate-400">AI-powered Terms-of-Service summaries</span>
        </div>
      </header>

      <main className="mx-auto flex max-w-5xl flex-col gap-6 px-6 py-10">
        <section className="rounded-2xl border border-slate-800 bg-slate-900/70 p-8 shadow-lg shadow-slate-900/40">
          <h1 className="text-3xl font-bold tracking-tight text-slate-50">Welcome to Auto-ToS</h1>
          <p className="mt-3 max-w-3xl text-lg text-slate-300">
            This is a minimal scaffold for the AI-powered Terms-of-Service summarizer. The search
            experience and admin tools will be built in the next steps. For now, the frontend is wired
            with Tailwind CSS and ready to consume the backend API.
          </p>
          <div className="mt-6 grid gap-4 md:grid-cols-2">
            <div className="rounded-xl border border-slate-800 bg-slate-900/60 p-5">
              <h2 className="text-xl font-semibold text-slate-50">What&apos;s next?</h2>
              <ul className="mt-3 space-y-2 text-sm text-slate-300">
                <li className="flex items-start gap-2">
                  <span className="mt-1 h-2 w-2 rounded-full bg-emerald-400" aria-hidden />
                  Build the search bar and summary display components.
                </li>
                <li className="flex items-start gap-2">
                  <span className="mt-1 h-2 w-2 rounded-full bg-emerald-400" aria-hidden />
                  Connect to the backend summary endpoints once available.
                </li>
                <li className="flex items-start gap-2">
                  <span className="mt-1 h-2 w-2 rounded-full bg-emerald-400" aria-hidden />
                  Add reporting and admin editing flows for summaries.
                </li>
              </ul>
            </div>
            <div className="rounded-xl border border-slate-800 bg-slate-900/60 p-5">
              <h2 className="text-xl font-semibold text-slate-50">Dev quickstart</h2>
              <ul className="mt-3 space-y-2 text-sm text-slate-300">
                <li>Run <code className="rounded bg-slate-800 px-2 py-1">npm run dev</code> to start the frontend.</li>
                <li>Backend healthcheck is served from <code className="rounded bg-slate-800 px-2 py-1">/healthz</code>.</li>
                <li>Docker Compose will wire the frontend, backend, and PostgreSQL together.</li>
              </ul>
            </div>
          </div>
        </section>
      </main>
    </div>
  );
}

export default App;
