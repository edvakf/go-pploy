<script>
  import { onDestroy, onMount } from 'svelte';
  import Commits from './Commits.svelte';

  export let status;

  let commandLog;
  let commandLogFrame;
  let commitLogFrame;

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
    const interval = setInterval(() => {
      if (commandLogFrame.classList.contains('loading') && commandLogFrame.contentDocument.body) {
        commandLogFrame.contentWindow.scrollTo(0, commandLogFrame.contentDocument.body.offsetHeight);
      }
    }, 200);

    return () => clearInterval(interval);
  });

  let submitted = false; // the form is submitted into the command log frame at least once. this prevents doneCommand to be called on page load.

  function submitCommandForm() {
    submitted = true;

    commandLog.classList.remove('hidden');

    // set border to red
    commandLogFrame.classList.add('loading');

    disableAllButtons();
  }

  function doneCommand() {
    if (!submitted) {
      return;
    }

    // set border to back to black
    commandLogFrame.classList.remove('loading');

    loadCommits();

    enableAllButtons();
  }

  let commits = null;

  function loadCommits() {
    if (commits) {
      commits.$destroy(); // without this, the commits iframe won't be reloaded.
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
  class="card {status.currentProject.lock && status.currentProject.lock.user === status.currentUser ? 'border-danger' : 'border-primary'}">
  <div class="card-header">{status.currentProject.name}</div>

  {#if status.currentProject.lock && status.currentProject.lock.user === status.currentUser}
    <div class="p-4 bg-light">
      <h5>Checkout</h5>
      <form action="./{status.currentProject.name}/checkout" method="post" class="form-inline command-form" target="command-log-frame" on:submit="{submitCommandForm}">
        <input type="text" class="form-control" name="ref" value="origin/{status.currentProject.defaultBranch}" required>
        <button class="btn btn-success checkout-button">Checkout</button>
      </form>
      <form action="./{status.currentProject.name}/deploy" method="post" class="command-form" target="command-log-frame" on:submit="{submitCommandForm}">
        {#each status.currentProject.deployEnvs as env}
          <h5>Deploy to {env}</h5>
          <button class="btn btn-success deploy-button" name="target" disabled value="{env}">Deploy to {env}</button>
        {/each}
      </form>
    </div>
  {/if}

  <div class="card-body">
    {#if status.currentProject.readme}
      <div class="card-text">{@html status.currentProject.readme}</div>
    {/if}

    <div id="command-log" class="hidden embed-responsive embed-responsive-16by9" bind:this={commandLog}>
      <iframe name="command-log-frame" class="log-frame embed-responsive-item" src="about:blank" bind:this={commandLogFrame} on:load="{doneCommand}" title="commit logs"></iframe>
    </div>

    <h4 class="p-1">Recent Commits</h4>
    <div id="commit-log" class="embed-responsive embed-responsive-16by9">
      <iframe class="log-frame embed-responsive-item" src="./assets/commits.html" on:load="{loadCommits}" bind:this={commitLogFrame} title="recent commits"></iframe>
    </div>

    <h4 class="p-1">Previous log <a href="./{status.currentProject.name}/logs?full=1&amp;generation=0" target="_blank">&#x27a1;</a></h4>
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