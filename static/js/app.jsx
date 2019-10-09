class App extends React.Component {
    render() {
        return (<Home />);
    }
  }
  
  class Home extends React.Component {
    render() {
      return (
        <div>
            <NavBar/>
            <div className="container">
                <div className="col-xs-8 col-xs-offset-2 jumbotron text-center">
                    <h1>Jokeish</h1>
                    <p>A load of Dad jokes XD</p>
                    <p>Sign in to get access </p>
                    <a  className="btn btn-primary btn-lg btn-login btn-block">Sign In</a>
                </div>
            </div>
        </div>
      )
    }
  }

  class NavBar extends React.Component {
      render() {
          return (
            <nav class="navbar navbar-dark fixed-top bg-dark flex-md-nowrap p-0 shadow">
            <ul class="navbar-nav px-3">
              <li class="nav-item active">
                <a class="nav-link" href="#">Home <span class="sr-only">(current)</span></a>
              </li>
              <li class="nav-item">
                <a class="nav-link" href="#">Link</a>
              </li>
              <li class="nav-item">
                <a class="nav-link disabled" href="#">Disabled</a>
              </li>
            </ul>
            </nav>
          )
      }
  }

  ReactDOM.render(<App />, document.getElementById('app'));
