import './index.css';

function App() {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      <header className="border-b border-slate-800 bg-slate-900/60 backdrop-blur">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
          <div className="text-lg font-bold tracking-tight">TOSAI</div>
          <span className="text-sm text-slate-400">Résumés automatiques et notés des CGU</span>
        </div>
      </header>

      <main className="mx-auto flex max-w-5xl flex-col gap-6 px-6 py-10">
        <section className="rounded-2xl border border-slate-800 bg-slate-900/70 p-8 shadow-lg shadow-slate-900/40">
          <h1 className="text-3xl font-bold tracking-tight text-slate-50">Bienvenue sur TOSAI</h1>
          <p className="mt-3 max-w-3xl text-lg text-slate-300">
            Ce squelette React prépare l&apos;interface utilisateur de TOSAI : récupérer des CGU/ToS,
            les résumer avec l&apos;IA, attribuer une note de A à E et appliquer des quotas par domaine.
            Les composants de recherche, de signalement et le panneau admin seront branchés sur l&apos;API
            au fur et à mesure.
          </p>
          <div className="mt-6 grid gap-4 md:grid-cols-2">
            <div className="rounded-xl border border-slate-800 bg-slate-900/60 p-5">
              <h2 className="text-xl font-semibold text-slate-50">Prochaines étapes</h2>
              <ul className="mt-3 space-y-2 text-sm text-slate-300">
                <li className="flex items-start gap-2">
                  <span className="mt-1 h-2 w-2 rounded-full bg-emerald-400" aria-hidden />
                  Construire la recherche par domaine et l&apos;affichage des résumés notés.
                </li>
                <li className="flex items-start gap-2">
                  <span className="mt-1 h-2 w-2 rounded-full bg-emerald-400" aria-hidden />
                  Connecter les composants aux endpoints de synthèse IA du backend.
                </li>
                <li className="flex items-start gap-2">
                  <span className="mt-1 h-2 w-2 rounded-full bg-emerald-400" aria-hidden />
                  Ajouter le bouton de signalement et les workflows admin sans comptes publics.
                </li>
              </ul>
            </div>
            <div className="rounded-xl border border-slate-800 bg-slate-900/60 p-5">
              <h2 className="text-xl font-semibold text-slate-50">Démarrage dev</h2>
              <ul className="mt-3 space-y-2 text-sm text-slate-300">
                <li>Lancez <code className="rounded bg-slate-800 px-2 py-1">npm run dev</code> pour démarrer le frontend.</li>
                <li>La santé backend est disponible sur <code className="rounded bg-slate-800 px-2 py-1">/healthz</code>.</li>
                <li>Docker Compose relie frontend, backend et PostgreSQL automatiquement.</li>
              </ul>
            </div>
          </div>
        </section>
      </main>
    </div>
  );
}

export default App;
