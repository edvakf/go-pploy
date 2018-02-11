import App from './components/App.js';

const pathComponents = location.pathname.split('/');
// const pathPrefix = pathComponents.length === 2 ? null : pathComponents[pathComponents.length - 2];
const project = pathComponents[pathComponents.length - 1];

fetch(`./api/status/${project}`).then(function(response) {
  return response.json();
}).then(function(status) {

  const app = new App({
    target: document.querySelector('main'),
    data: {
      status: status,
    },
  });

}).catch(function(error) {
  console.log(error);
  alert('error fetching the status api');
});
