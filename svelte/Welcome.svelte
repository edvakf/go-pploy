<script>
  export let status;
</script>

<h2>Welcome to pploy</h2>

<p>Click on a project in the sidebar.</p>

<h3 class="p-1">Add a project</h3>

<form action="./_create" method="post">
  <div class="input-group">
    <input type="text" class="form-control" name="url" value="" placeholder="https://github.com/edvakf/testploy.git"
      required>
    <span class="input-group-btn">
      <button class="btn btn-success" type="submit">git clone</button>
    </span>
  </div>
</form>