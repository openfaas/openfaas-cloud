import React, { Component } from 'react';
import { BrowserRouter, Route, Switch } from 'react-router-dom';

import './App.css';

import { NavBar } from './components/NavBar';
import { FunctionsOverviewPage } from './pages/FunctionsOverviewPage';
import { FunctionDetailPage } from './pages/FunctionDetailPage';
import { FunctionLogPage } from './pages/FunctionLogPage';
import { NotFoundPage } from './pages/NotFoundPage';
import { Breadcrumbs } from './components/Breadcrumbs';

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
          <p>
            Powered by <a href="https://www.openfaas.com">OpenFaaS</a>
          </p>
        </div>
      </BrowserRouter>
    );
  }
}
