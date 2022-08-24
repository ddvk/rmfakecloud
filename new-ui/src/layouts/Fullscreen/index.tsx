import { Outlet } from 'react-router-dom'

export default function Fullscreen() {
  return (
    <main className="grid min-h-screen bg-slate-900">
      <Outlet />
    </main>
  )
}
