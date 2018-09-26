import React, { Component } from 'react';
import { BrowserRouter, Route, Switch } from 'react-router-dom';

import { NavBar } from './components/NavBar';
import { FunctionsOverviewPage } from './pages/FunctionsOverviewPage';
import { FunctionDetailPage } from './pages/FunctionDetailPage';
import { FunctionLogPage } from './pages/FunctionLogPage';
import { NotFoundPage } from './pages/NotFoundPage';
import { Breadcrumbs } from './components/Breadcrumbs';
import { Footer } from './components/Footer';

import 'bootstrap/dist/css/bootstrap.min.css';
import './App.css';

export class App extends Component {
  render() {
    // basename is injected from the server
    const basename =
      process.env.NODE_ENV === 'production' ? window.BASE_HREF : '/';

    return (
      <BrowserRouter basename={basename}>
        <div className="container">
          <NavBar />
          <Breadcrumbs />
          <div>
            <Switch>
              <Route exact path="/:user" component={FunctionsOverviewPage} />
              <Route
                exact
                path="/:user/:functionName"
                component={FunctionDetailPage}
              />
              <Route
                path="/:user/:functionName/log"
                component={FunctionLogPage}
              />
              <Route component={NotFoundPage} />
            </Switch>
          </div>
          <Footer />
        </div>
      </BrowserRouter>
    );
  }
}
