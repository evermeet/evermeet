<script>
  /**
   * Copyright (c) Meta Platforms, Inc. and affiliates.
   *
   * This source code is licensed under the MIT license found in the
   * LICENSE file in the root directory of this source tree.
   *
   */

  import { registerDragonSupport } from "@lexical/dragon";
  import { createEmptyHistoryState, registerHistory } from "@lexical/history";
  import { HeadingNode, QuoteNode, registerRichText } from "@lexical/rich-text";
  import { mergeRegister } from "@lexical/utils";
  import { createEditor } from "lexical";

  import { onMount } from "svelte";

  import {
    $createHeadingNode as _createHeadingNode,
    $createQuoteNode as _createQuoteNode,
  } from "@lexical/rich-text";
  import {
    $createParagraphNode as _createParagraphNode,
    $createTextNode as _createTextNode,
    $getRoot as _getRoot,
  } from "lexical";

  function prepopulatedRichText() {
    const root = _getRoot();
    if (root.getFirstChild() !== null) {
      return;
    }

    const heading = _createHeadingNode("h1");
    heading.append(_createTextNode("Welcome to the Vanilla JS Lexical Demo!"));
    root.append(heading);
    const quote = _createQuoteNode();
    quote.append(
      _createTextNode(
        `In case you were wondering what the text area at the bottom is – it's the debug view, showing the current state of the editor. `,
      ),
    );
    root.append(quote);
    const paragraph = _createParagraphNode();
    paragraph.append(
      _createTextNode("This is a demo environment built with "),
      _createTextNode("lexical").toggleFormat("code"),
      _createTextNode("."),
      _createTextNode(" Try typing in "),
      _createTextNode("some text").toggleFormat("bold"),
      _createTextNode(" with "),
      _createTextNode("different").toggleFormat("italic"),
      _createTextNode(" formats."),
    );
    root.append(paragraph);
  }
  let editorRef;
  let stateRef;

  const initialConfig = {
    namespace: "Vanilla JS Demo",
    // Register nodes specific for @lexical/rich-text
    nodes: [HeadingNode, QuoteNode],
    onError: (error) => {
      throw error;
    },
    theme: {
      // Adding styling to Quote node, see styles.css
      quote: "PlaygroundEditorTheme__quote",
    },
  };
  onMount(() => {
    const editor = createEditor(initialConfig);
    editor.setRootElement(editorRef);

    // Registering Plugins
    mergeRegister(
      registerRichText(editor),
      registerDragonSupport(editor),
      registerHistory(editor, createEmptyHistoryState(), 300),
    );

    editor.update(prepopulatedRichText, { tag: "history-merge" });

    editor.registerUpdateListener(({ editorState }) => {
      stateRef.value = JSON.stringify(editorState.toJSON(), undefined, 2);
    });
  });
</script>

<h1>SvelteKit Lexical Basic - Vanilla JS</h1>
<div class="editor-wrapper">
  <div id="lexical-editor" bind:this={editorRef} contenteditable></div>
</div>
<h4>Editor state:</h4>
<textarea id="lexical-state" bind:this={stateRef}></textarea>

<style>
  .editor-wrapper {
    border: 2px solid gray;
  }
  #lexical-state {
    width: 100%;
    height: 300px;
  }
  :global(.PlaygroundEditorTheme__quote) {
    margin: 0;
    margin-left: 20px;
    margin-bottom: 10px;
    font-size: 15px;
    color: rgb(101, 103, 107);
    border-left-color: rgb(206, 208, 212);
    border-left-width: 4px;
    border-left-style: solid;
    padding-left: 16px;
  }
</style>
