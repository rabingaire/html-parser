import React, { useState } from 'react';

import { getPageInfo } from './service';

import './App.scss';

function App() {
  const [searchText, setSearchText] = useState('');
  const [loading, setLoading] = useState(false);
  const [info, setInfo] = useState(null);

  const handleOnSearchFormChange = (event) => {
    setSearchText(event.target.value);
  };

  const handleOnKeyDown = (event) => {
    if (event.key === 'Enter') {
      handleOnSearchClick();
    }
  };

  const handleOnSearchClick = () => {
    setLoading(true);
    setInfo(null);
    getPageInfo(searchText)
      .then((data) => {
        setLoading(false);
        setInfo(data);
      })
      .catch((err) => {
        setLoading(false);
        console.error(err);
      });
  };

  return (
    <div className="container">
      <h1>Page Info</h1>
      <div className="container__search">
        <input
          type="text"
          value={searchText}
          onChange={handleOnSearchFormChange}
          onKeyDown={handleOnKeyDown}
          placeholder="https://example.com"
        />
        <button onClick={handleOnSearchClick} disabled={loading}>
          <span>Search</span>
        </button>
      </div>
      {info ? (
        <div className="container__result">
          <pre>{JSON.stringify(info, null, 2)}</pre>
        </div>
      ) : null}
    </div>
  );
}

export default App;
