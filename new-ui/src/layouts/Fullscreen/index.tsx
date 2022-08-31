import { Outlet } from 'react-router-dom'

export default function Fullscreen() {
  return (
    <main className="grid min-h-screen bg-gradient-to-bl from-slate-700 via-slate-900 to-slate-800">
      <Outlet />
    </main>
  )
}
