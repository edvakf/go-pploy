<script>
  import { onMount } from 'svelte';

  import Lock from './Lock.svelte';
  import Projects from './Projects.svelte';
  import Main from './Main.svelte';
  import Welcome from './Welcome.svelte';
  import Remove from './Remove.svelte';

  export let status;

  onMount(() => {
    const pathComponents = location.pathname.split('/');
    const project = pathComponents[pathComponents.length - 1];

    fetchStatusAPI(project);

    setInterval(() => {
      fetchStatusAPI(project);
    }, 10000);
  });

  function fetchStatusAPI(project) {
    fetch(
      `./api/status/${project}`,
      {
        credentials: 'same-origin',
      }
    ).then((response) => {
      return response.json();
    }).then((_status) => {
      if (_status.message !== "") {
        iziToast.error({ message: _status.message, position: 'topRight' });
      }
      status = _status;
    }).catch((error) => {
      console.log(error);
      iziToast.error({ message: error.message, position: 'topRight' });
    });
  }
</script>

<nav class="navbar navbar-light bg-light mb-3">
  <div class="container">
    <a class="navbar-brand" href="./">pploy</a>
  </div>
</nav>

<div class="container">
  <div class="row">

    {#if status}

      <!-- main -->
      <div class="col-md-9">
        {#if status.currentProject}
        <Main {status}></Main>
        {:else}
        <Welcome {status}></Welcome>
        {/if}
      </div>

      <!-- sidebar -->
      <aside class="col-md-3">
        {#if status.currentProject}
        <Lock {status}></Lock>
        {/if}

        <Projects {status}></Projects>

        {#if status.currentProject}
        <Remove {status}></Remove>
        {/if}
      </aside>

    {/if}

  </div>
</div>