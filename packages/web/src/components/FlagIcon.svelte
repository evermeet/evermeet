<script>
  //const icons = import.meta.glob(`../../node_modules/flag-icons/flags/1x1/*.svg`, { eager: true })

  export let country;
  export let countryName;
  export let size = "1em";
</script>

<span
  alt={countryName}
  title={countryName}
  class="fi fi-{country} fis rounded-full"
  style="line-height: {size}; width: {size};"
></span>
