<script>
  import { onMount } from 'svelte';

  export let commits = [];
  export let project;

  onMount(() => {

  const fetchStatusAPI = fetch(
    `./api/commits/${project.name}`,
    {
      credentials: 'same-origin',
    }
  ).then((response) => {
    return response.json();
  }).then((commits) => {
    commits = commits;
  }).catch((error) => {
    message = error.message;
  });
});
</script>

<table>
  <tbody>
    {#each commits as commit}
      <tr class="header">
        <td class="hash">
          {commit.hash.slice(0, 7)}
        </td>
        <td nowrap>{commit.author}</td>
        <td>
          {#each commit.otherRefs as ref}
            {#if ref === "HEAD"}
              <span class="ref head">HEAD</span>
            {:elseif ref.startsWith("refs/remotes/origin/master")}
              <span class="ref master">origin/master</span>
            {:elseif ref.startsWith("refs/remotes/")}
              <span class="ref">{ref.slice("refs/remotes/".length)}</span>
            {:elseif ref.startsWith("refs/heads/")}
            {:elseif ref.startsWith("refs/tags/")}
              <span class="ref tag">{ref.slice("refs/tags/".length)}</span>
            {:elseif ref.startsWith("tag: refs/tags/")}
              <span class="ref tag">{ref.slice("tag: refs/tags/".length)}</span>
            {:else}
              <span class="ref">{ref}</span>
            {/if}
          {/each}
        </td>
        <td nowrap>{commit.time}</td>
      </tr>
      <tr class="subject">
        <td colspan="5">{commit.subject}</td>
      </tr>
      <tr>
        <td class="code" colspan="5">
        <pre>{commit.nameStatus}</pre>
        </td>
      </tr>
    {/each}
  </tbody>
</table>
