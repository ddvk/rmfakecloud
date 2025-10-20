import Container from "react-bootstrap/Container";

const Home = () => {
  return (
    <Container fluid>
      <main>
        <h1>Welcome to your own reMarkable Cloud!</h1>
        <h2>About</h2>
        <p>
          This software is an unofficial replacement for the proprietary
          reMarkable Cloud.  In case you want to sync/backup your files
          and have full control of the hosting environment, this is the
          software for you.
        </p>
        <p>
          It's is still a work in progress being, actively maintained over
          on <a href="https://github.com/ddvk/rmfakecloud">GitHub</a>.
        </p>
        <h2>Tips</h2>
        <ul>
          <li>
            <p>
              You can use <a href="https://github.com/juruen/rmapi">rmapi</a> for managing files,
              just specify the URL of your instance with the RMAPI_HOST variable like
              so: <code>RMAPI_HOST=https//rmfakeclud.example.com rmapi</code>
            </p>
            <ul>
              <li>
                <p>
                  Do note that the original project is now unmaintained. You should consider
                  using <a href="https://github.com/ddvk/rmapi">this fork</a> instead.
                </p>
              </li>
            </ul>
          </li>
          <li>
            <p>
              Check out the online <a href="https://ddvk.github.io/rmfakecloud/">documentation</a> to
              learn more about the configuration options. Also read the README
            </p>
          </li>
          <li>
            <p>
              You should also read the <a href="https://github.com/ddvk/rmfakecloud/blob/master/README.md">README</a>,
              to see the current status of the project and notes from the developers.
            </p>
          </li>
          <li>
            <p>
              We support the Read on reMarkable Extension. Read more about it in the online documentation.
            </p>
          </li>
          <li>
            <p>
              Documents will be uploaded to the selected (highlighted) directory.
            </p>
            <ul>
              <li>
                <p>
                  Select directories by clicking on them.
                </p>
              </li>
              <li>
                <p>
                  Clicking on selected directories again will open or close them respectively.
                </p>
              </li>
            </ul>
          </li>
        </ul>
      </main>
    </Container>
  );
};

export default Home;
