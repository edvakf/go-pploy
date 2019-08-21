<script>
  export let status;
</script>

<form action="./{status.currentProject.name}/remove" method="post" class="sidebar-section bg-light p-3 mb-3">
  <h4>Remove project</h4>

  <p>Please lock this project in order to remove.</p>

  <button class="btn btn-danger" disabled={!status.currentProject.lock || status.currentProject.lock.user !== status.currentUser}>
    &#x2620; Remove this project
  </button>
</form>
