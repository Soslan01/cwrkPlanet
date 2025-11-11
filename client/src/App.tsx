import { TopBar } from "./components/TopBar";
import { Dock } from "./components/Dock";
import { useHashRoute } from "./lib/router";
import { Home } from "./pages/Home";
import { Auth } from "./pages/Auth";
import { Profile } from "./pages/Profile";
import RoomsPage from "./pages/Rooms";
import RoomPage from "./pages/Room";

export default function App() {
  const { route, params } = useHashRoute();

  let Page: React.ComponentType<any> = Home;
  let pageProps: any = {};

  if (route.startsWith("/auth")) Page = Auth;
  else if (route.startsWith("/profile")) Page = Profile;
  else if (route === "/rooms") Page = RoomsPage;
  else {
    const m = route.match(/^\/rooms\/([^/]+)$/);
    if (m) { Page = RoomPage; pageProps = { roomId: params.roomId! }; }
  }

  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-900 dark:bg-neutral-950 dark:text-neutral-100 transition-colors">
      <TopBar />
      {}
      <main className="mx-auto max-w-3xl px-4 pt-20 pb-[var(--dock-height)]">
        <Page {...pageProps} />
      </main>
      <Dock />
    </div>
  );
}
