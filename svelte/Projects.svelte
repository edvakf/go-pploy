<script>
  export let status;
</script>

<div class="sidebar-section bg-light p-3 mb-3">
  <h4>Projects</h4>

  <ul class="nav nav-pills flex-column">
    {#each status.allProjects as project}
      <li class="nav-item" >
        <a href="./{project.name}" class="nav-link {status.currentProject && status.currentProject.name === project.name ? 'active' : ''}">
          {project.name}
          {#if project.lock}&#x23f3;{/if}
        </a>
      </li>
    {/each}
  </ul>
</div>
