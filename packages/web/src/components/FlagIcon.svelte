<script>
    const icons = import.meta.glob(`../../node_modules/flag-icons/flags/1x1/*.svg`, { eager: true })

    export let country;
    export let size = "w-3.5 h-3.5";

    let icon = icons[`../../node_modules/flag-icons/flags/1x1/${country}.svg`].default;

</script>

<img src={icon} alt={country} class="{size} rounded-full aspect-square" />