<script>
  import { user } from "$lib/stores";
  import { goto } from "$app/navigation";

  user.set(false);
  goto("/");
</script>
