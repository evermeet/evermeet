<script>
    export let placeholder;
    export let str;

    const lineHeight = 52;
    const origHeight = 48;
    let lines = 1;
    let backspace = false;

    str.subscribe((current) => {
        if (current.length > 0) {
            backspace = false;
            const scrollHeight = document.getElementById('my-textarea').scrollHeight;
            if (scrollHeight === lineHeight) {
                lines = 1
            }
            else if (scrollHeight % lineHeight !== 0) {
                lines = ((scrollHeight-lineHeight)/origHeight) + 1;
            }
        }
        str.set(current.replace(/[^\x20-\x7E]/gmi, ""))
    })

    function onKeyPress (event) {
        /*if (event.keyCode == 13) {
            event.preventDefault();
        }*/
        if (event.keyCode == 8) {
            backspace = true;
        }
    }
    

</script>

<div class="grow-wrap" style="height: {lines*lineHeight}px; width:100%;">
    <textarea id="my-textarea" rows="1" type="text"
        class="my-textarea {$str.length > 0 ? "active" : ""}" {placeholder}
        bind:value={$str}
        on:keydown={onKeyPress}
        style="height: {backspace ? 'auto' : (lineHeight*lines)+'px'}; width: 100%;"
    ></textarea>
</div>

<style>
    .my-textarea {
        @apply text-5xl font-mono font-semibold bg-transparent outline-0 resize-none opacity-75 hover:opacity-100 h-full
    }
    .my-textarea.active {
        @apply opacity-100;
    }
</style>