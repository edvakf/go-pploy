<script>
  export let status;

  import { onDestroy, onMount } from 'svelte';

  export let commandLog;
  export let commandLogFrame;
  export let commitLogFrame;

  import Commits from './Commits.svelte';

  function disableAllButtons() {
    setTimeout(function () {
      Array.from(document.querySelectorAll('button'))
        .forEach(b => b.setAttribute('disabled', 'disabled'));
    }, 10);
  }

  function enableAllButtons() {
    setTimeout(function () {
      Array.from(document.querySelectorAll('button'))
        .forEach(b => b.removeAttribute('disabled'));
    }, 10);
  }

  onMount(() => {
    // follow scroll
    const f = commandLogFrame;
    const interval = setInterval(() => {
      if (f.classList.contains('loading')) {
        f.contentWindow.scrollTo(0, f.contentDocument.body.offsetHeight);
      }
    }, 200);

    return () => clearInterval(interval);
  });

  function submitCommandForm() {
    commandLog.classList.remove('hidden');

    // set border to red
    commandLogFrame.classList.add('loading');

    disableAllButtons();
  }

  function doneCommand() {
    if (!commandLogFrame) {
      return; // on:load is called while dom is still incomplete
    }

    // set border to back to black
    commandLogFrame.classList.remove('loading');

    loadCommits();

    enableAllButtons();
  }

  let commits = null;

  function loadCommits() {
    if (commits) {
      commits.destroy();
    }
    commits = new Commits({
      target: commitLogFrame.contentDocument.querySelector('commits'),
      props: {
        project: status.currentProject,
      },
    });
  }
</script>

<div
  class="panel {status.currentProject.lock && status.currentProject.lock.user === status.currentUser ? 'panel-danger' : 'panel-primary'}">
  <div class="panel-heading">
    <h2 class="panel-title">{status.currentProject.name}</h2>
  </div>

  {#if status.currentProject.lock && status.currentProject.lock.user === status.currentUser}
    <div class="well">
      <h4>Checkout</h4>
      <form action="./{status.currentProject.name}/checkout" method="post" class="form-inline command-form" target="command-log-frame" on:submit="{submitCommandForm}">
        <input type="text" class="form-control" name="ref" value="origin/{status.currentProject.defaultBranch}" required>
        <button class="btn btn-success checkout-button">Checkout</button>
      </form>
      <form action="./{status.currentProject.name}/deploy" method="post" class="command-form" target="command-log-frame" on:submit="{submitCommandForm}">
        {#each status.currentProject.deployEnvs as env}
          <h4>Deploy to {env}</h4>
          <button class="btn btn-success deploy-button" name="target" disabled value="{env}">Deploy to {env}</button>
        {/each}
      </form>
    </div>
  {/if}

  <div class="panel-body">
    {#if status.currentProject.readme}
      <div>{@html status.currentProject.readme}</div>
    {/if}

    <div id="command-log" class="hidden embed-responsive embed-responsive-16by9" bind:this={commandLog}>
      <iframe name="command-log-frame" class="log-frame embed-responsive-item" src="about:blank" bind:this={commandLogFrame} on:load="{doneCommand}" title="commit logs"></iframe>
    </div>

    <h3>Recent Commits</h3>
    <div id="commit-log" class="embed-responsive embed-responsive-16by9">
      <iframe class="log-frame embed-responsive-item" src="./assets/commits.html" on:load="{loadCommits}" bind:this={commitLogFrame} title="recent commits"></iframe>
    </div>

    <h3>Previous log <a href="./{status.currentProject.name}/logs?full=1&amp;generation=0" target="_blank" class="glyphicon glyphicon-hand-right"></a></h3>
    <div class="embed-responsive embed-responsive-16by9">
      <iframe class="log-frame embed-responsive-item" src="./{status.currentProject.name}/logs" title="previous logs"></iframe>
    </div>
  </div>
</div>

<style>
.log-frame {
  border: solid black 2px !important; /* override twitter bootstrap's embed-responsive iframe */
}
.log-frame.loading {
  transition-duration: 0.4s;
  transition-property: border;
  border: solid red 2px !important;
}
</style>