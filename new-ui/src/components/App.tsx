import react from '@/Assets/images/react.svg'
import tailwindcss from '@/Assets/images/tailwindcss.svg'
import typescript from '@/Assets/images/typescript.svg'
import vercel from '@/Assets/images/vercel.svg'
import vite from '@/Assets/images/vite.svg'

function App() {
  return (
    <main className="grid min-h-screen place-content-center bg-gradient-to-b from-blue-700 to-blue-800">
      <section className="flex flex-col items-center justify-center gap-7 text-center text-blue-100">
        <h1 className="text-7xl font-bold tracking-wide">
          VRTTV
          <span className="block text-3xl italic">Boilerplate</span>
        </h1>
        <p className="max-w-sm text-base leading-7 sm:max-w-none">
          Avoid setting up a project from scratch. Start using VRTTV 🎉
        </p>
        <a
          className="rounded bg-blue-100 py-3 px-4 font-bold uppercase tracking-wide text-blue-700 shadow-md shadow-blue-800 transition-colors hover:bg-blue-900 hover:text-blue-100"
          href="https://github.com/Drumpy/vrttv-boilerplate"
          rel="noopener noreferrer"
          target="_blank"
        >
          Get the boilerplate →
        </a>
        <div className="flex gap-8 pt-4">
          <img
            alt="Vite Icon"
            className="text-blue-200 hover:text-blue-100"
            height="32px"
            src={vite}
            width="32px"
          />
          <img
            alt="React Icon"
            className="fill-blue-500 hover:text-blue-100"
            height="32px"
            src={react}
            width="32px"
          />
          <img
            alt="Typescript Icon"
            className="fill-blue-500 hover:text-blue-100"
            height="32px"
            src={typescript}
            width="32px"
          />
          <img
            alt="Tailwindcss Icon"
            className="fill-blue-500 hover:text-blue-100"
            height="32px"
            src={tailwindcss}
            width="32px"
          />
          <img
            alt="Vercel Icon"
            className="fill-blue-500 hover:text-blue-100"
            height="32px"
            src={vercel}
            width="32px"
          />
        </div>
      </section>
    </main>
  )
}

export default App
