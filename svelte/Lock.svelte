<script>
  export let status;

  import { onDestroy, onMount } from 'svelte';

  function pad02(num) {
    return ('0' + num).substr(-2);
  }

  function secondsToString(seconds) {
    return pad02(Math.floor(seconds / 60)) + ':' + pad02(Math.floor(seconds % 60));
  }

  let now = Date.now();

  onMount(() => {
    const interval = setInterval(() => {
      now = Date.now();
    }, 1000);

    return () => clearInterval(interval);
  });

  const minutesAndSecondsLeft = (endTime, now) => {
    const timeLeft = Date.parse(endTime) - now;
    if (timeLeft < 0) {
      location.reload();
    }
    return secondsToString(timeLeft / 1000);
  };
</script>

<form class="sidebar-section bg-light p-3 mb-3" action="./{status.currentProject.name}/lock" method="POST" id="lock-form"
  data-lock-user="{status.currentProject.lock ? status.currentProject.lock.user : ''}">
  {#if status.currentProject.lock}
    <p>
      Working
      <span class="badge badge-secondary">{status.currentProject.lock.user}</span>
    </p>
    <p>
      Time left
      <span class="badge badge-danger time-left">
        {minutesAndSecondsLeft(status.currentProject.lock.endTime, now)}
      </span>
    </p>
    {#if status.currentProject.lock.user === status.currentUser}
      <button class="btn btn-warning btn-block" name="operation" value="extend">Extend</button>
      <button class="btn btn-success btn-block" name="operation" value="release">Finish deploying</button>
      <input type="hidden" name="user" value="{status.currentProject.lock.user}">
    {/if}
  {:else}
    <select class="form-control" name="user">
      <option value="">[Please Select]</option>
      {#each status.allUsers as user}
        <option value="{user}" selected="{user === status.currentUser}">
          {user}
        </option>
      {/each}
    </select>
    <button class="btn btn-success btn-block" name="operation" value="gain">Start deploying</button>
  {/if}
</form>
