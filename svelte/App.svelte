<script>
  import { onMount } from 'svelte';

  import Lock from './Lock.svelte';
  import Projects from './Projects.svelte';
  import Main from './Main.svelte';
  import Welcome from './Welcome.svelte';

  export let status;

  onMount(() => {
    const pathComponents = location.pathname.split('/');
    // const pathPrefix = pathComponents.length === 2 ? null : pathComponents[pathComponents.length - 2];
    const project = pathComponents[pathComponents.length - 1];

    fetchStatusAPI(project);

    setInterval(() => {
      fetchStatusAPI(project);
    }, 10000);
  });

  // [svelte-upgrade suggestion]
  // review these functions and remove unnecessary 'export' keywords
  export function fetchStatusAPI(project) {
    fetch(
      `./api/status/${project}`,
      {
        credentials: 'same-origin',
      }
    ).then((response) => {
      return response.json();
    }).then((status) => {
      if (status.message !== "") {
        iziToast.error({ message: status.message, position: 'topRight' });
      }
      status = status;
    }).catch((error) => {
      console.log(error);
      iziToast.error({ message: error.message, position: 'topRight' });
    });
  }
</script>

<header class="navbar navbar-default" role="navigation">
  <div class="container">
    <div class="navbar-header">
      <a class="navbar-brand" href="./">pploy</a>
    </div>
  </div>
</header>

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
      </aside>

    {/if}

  </div>
</div>