const host = process.env.REACT_APP_API_HOST || 'http://localhost:8000/api/v1';

const getPageInfo = (url) => {
  return new Promise((resolve, reject) => {
    fetch(`${host}/info?url=${url}`)
      .then((response) => response.json())
      .then((data) => resolve(data))
      .catch((err) => reject(err));
  });
};

export { getPageInfo };
