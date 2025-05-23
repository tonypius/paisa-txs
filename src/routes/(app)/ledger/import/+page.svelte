<script lang="ts">
  import Select from "svelte-select";
  import {
    createEditor as createTemplateEditor,
    editorState as templateEditorState
  } from "$lib/template_editor";
  import {
    createEditor as createPreviewEditor,
    updateContent as updatePreviewContent
  } from "$lib/editor";
  import Dropzone from "svelte-file-dropzone/Dropzone.svelte";
  import { parse, asRows, render as renderJournal } from "$lib/spreadsheet";
  import _ from "lodash";
  import type { EditorView } from "codemirror";
  import { onMount } from "svelte";
  import { ajax, type ImportTemplate } from "$lib/utils";
  import { accountTfIdf } from "../../../../store";
  import * as toast from "bulma-toast";
  import FileModal from "$lib/components/FileModal.svelte";
  import Modal from "$lib/components/Modal.svelte";

  let templates: ImportTemplate[] = [];
  let selectedTemplate: ImportTemplate;
  let saveAsName: string;
  let lastTemplate: any;
  let lastData: any;
  let preview = "";
  let parseErrorMessage: string = null;
  let columnCount: number;
  let data: any[][] = [];
  let rows: Array<Record<string, any>> = [];
  let lastOptions: any;
  let options: { reverse: boolean; trim: boolean } = { reverse: false, trim: true };
  let newTag: string = "";

  let templateEditorDom: Element;
  let templateEditor: EditorView;

  let previewEditorDom: Element;
  let previewEditor: EditorView;

  onMount(async () => {
    // Mock data for testing without backend
    accountTfIdf.set({});
    templates = [
      {
        id: "1",
        name: "Bank Statement",
        content: "{{#each ROW}}\n{{formatDate A \"YYYY-MM-DD\"}} {{B}}\n    {{C}}  {{D}} EUR\n{{/each}}",
        template_type: "custom",
        rules: [
          {
            condition: "ROW.B && ROW.B.includes('GROCERY')",
            tags: ["food"],
            skip: false
          },
          {
            condition: "ROW.D && parseFloat(ROW.D) > 1000",
            tags: ["large-expense"],
            skip: false
          }
        ]
      },
      {
        id: "2",
        name: "Credit Card",
        content: "{{#each ROW}}\n{{formatDate A \"YYYY-MM-DD\"}} {{B}}\n    {{C}}  {{D}} EUR\n{{/each}}",
        template_type: "custom",
        rules: []
      }
    ];
    
    if (templates && templates.length > 0) {
      selectedTemplate = templates[0];
      // Ensure rules is initialized
      if (!selectedTemplate.rules) {
        selectedTemplate.rules = [];
      }
      saveAsName = selectedTemplate.name;
      templateEditor = createTemplateEditor(selectedTemplate.content, templateEditorDom);
      previewEditor = createPreviewEditor(preview, previewEditorDom, { readonly: true });
    }
  });

  $: saveAsNameDuplicate = !!_.find(templates, { name: saveAsName, template_type: "custom" });

  async function save() {
    // Mock save function for testing without backend
    const newTemplate = {
      id: Date.now().toString(),
      name: saveAsName,
      content: templateEditor.state.doc.toString(),
      template_type: "custom",
      rules: selectedTemplate?.rules || []
    };
    
    // Check if we're updating an existing template
    const existingIndex = templates.findIndex(t => t.name === saveAsName && t.template_type === "custom");
    if (existingIndex >= 0) {
      // Update existing template
      templates[existingIndex] = {...newTemplate, id: templates[existingIndex].id};
    } else {
      // Add new template
      templates = [...templates, newTemplate];
    }
    
    selectedTemplate = _.find(templates, { name: saveAsName, template_type: "custom" });
    // Ensure rules is initialized
    if (!selectedTemplate.rules) {
      selectedTemplate.rules = [];
    }
    
    toast.toast({
      message: `Saved ${saveAsName}`,
      type: "is-success"
    });

    $templateEditorState = _.assign({}, $templateEditorState, { hasUnsavedChanges: false });
  }

  async function remove() {
    const oldName = selectedTemplate.name;
    const confirmed = confirm(`Are you sure you want to delete ${oldName} template?`);
    if (!confirmed) {
      return;
    }
    
    // Mock delete function for testing without backend
    templates = templates.filter(t => t.id !== selectedTemplate.id);
    
    if (templates && templates.length > 0) {
      selectedTemplate = templates[0];
      // Ensure rules is initialized
      if (!selectedTemplate.rules) {
        selectedTemplate.rules = [];
      }
      saveAsName = selectedTemplate.name;
    }
    toast.toast({
      message: `Removed ${oldName}`,
      type: "is-success"
    });

    $templateEditorState = _.assign({}, $templateEditorState, { hasUnsavedChanges: false });
  }

  let input: any;

  $: if (!_.isEmpty(data) && $templateEditorState.template) {
    if (
      lastTemplate != $templateEditorState.template ||
      lastData != data ||
      lastOptions != options
    ) {
      try {
        preview = renderJournal(rows, $templateEditorState.template, {
          reverse: options.reverse,
          trim: options.trim,
          rules: selectedTemplate?.rules || []
        });
        updatePreviewContent(previewEditor, preview);
        lastTemplate = $templateEditorState.template;
        lastData = data;
        lastOptions = _.clone(options);
      } catch (e) {
        console.log(e);
      }
    }
  }

  $: if (selectedTemplate && templateEditor) {
    if (templateEditor.state.doc.toString() != selectedTemplate.content) {
      templateEditor.destroy();
      templateEditor = createTemplateEditor(selectedTemplate.content, templateEditorDom);
    }
  }

  async function handleFilesSelect(e: { detail: { acceptedFiles: File[] } }) {
    const { acceptedFiles } = e.detail;

    const results = await parse(acceptedFiles[0]);
    if (results.error) {
      parseErrorMessage = results.error;
    } else {
      parseErrorMessage = null;
      data = results.data;
      rows = asRows(results);

      columnCount = _.maxBy(data, (row) => row.length).length;
      _.each(data, (row) => {
        row.length = columnCount;
      });
    }
  }

  async function copyToClipboard() {
    try {
      await navigator.clipboard.writeText(preview);
      toast.toast({
        message: "Copied to clipboard",
        type: "is-success"
      });
    } catch (e) {
      console.log(e);
      toast.toast({
        message: "Failed to copy to clipboard",
        type: "is-danger"
      });
    }
  }

  let modalOpen = false;
  function openSaveModal() {
    modalOpen = true;
  }

  async function saveToFile(destinationFile: string) {
    // Mock saveToFile function for testing without backend
    toast.toast({
      message: `Saved <b><a href="/ledger/editor/${encodeURIComponent(
        destinationFile
      )}">${destinationFile}</a></b>`,
      type: "is-success",
      duration: 5000
    });
  }

  function builtinNotAllowed(action: string, template: ImportTemplate) {
    if (template?.template_type == "builtin") {
      return `Not allowed to ${action.toLowerCase()} builtin template`;
    }
    return action;
  }

  let templateCreateModalOpen = false;
  function openTemplateCreateModal() {
    templateCreateModalOpen = true;
  }
  
  // Rules management
  function addRule() {
    if (!selectedTemplate) return;
    
    if (!selectedTemplate.rules) {
      selectedTemplate.rules = [];
    }
    selectedTemplate.rules.push({
      condition: "",
      tags: [],
      skip: false
    });
    selectedTemplate = {...selectedTemplate}; // Ensure reactivity
    $templateEditorState = _.assign({}, $templateEditorState, { hasUnsavedChanges: true });
  }
  
  function removeRule(index: number) {
    if (!selectedTemplate) return;
    
    if (selectedTemplate.rules && selectedTemplate.rules.length > index) {
      selectedTemplate.rules.splice(index, 1);
      selectedTemplate = {...selectedTemplate}; // Trigger reactivity
      $templateEditorState = _.assign({}, $templateEditorState, { hasUnsavedChanges: true });
    }
  }
  
  function updateRuleCondition(index: number, condition: string) {
    if (!selectedTemplate) return;
    
    if (selectedTemplate.rules && selectedTemplate.rules.length > index) {
      selectedTemplate.rules[index].condition = condition;
      $templateEditorState = _.assign({}, $templateEditorState, { hasUnsavedChanges: true });
    }
  }
  
  function updateRuleSkip(index: number, skip: boolean) {
    if (!selectedTemplate) return;
    
    if (selectedTemplate.rules && selectedTemplate.rules.length > index) {
      selectedTemplate.rules[index].skip = skip;
      $templateEditorState = _.assign({}, $templateEditorState, { hasUnsavedChanges: true });
    }
  }
  
  function addTag(index: number, tag: string) {
    if (!selectedTemplate) return;
    
    if (selectedTemplate.rules && selectedTemplate.rules.length > index) {
      if (!selectedTemplate.rules[index].tags) {
        selectedTemplate.rules[index].tags = [];
      }
      if (tag && !selectedTemplate.rules[index].tags.includes(tag)) {
        selectedTemplate.rules[index].tags.push(tag);
        selectedTemplate = {...selectedTemplate}; // Trigger reactivity
        $templateEditorState = _.assign({}, $templateEditorState, { hasUnsavedChanges: true });
      }
    }
  }
  
  function removeTag(ruleIndex: number, tagIndex: number) {
    if (!selectedTemplate) return;
    
    if (selectedTemplate.rules && 
        selectedTemplate.rules.length > ruleIndex && 
        selectedTemplate.rules[ruleIndex].tags && 
        selectedTemplate.rules[ruleIndex].tags.length > tagIndex) {
      selectedTemplate.rules[ruleIndex].tags.splice(tagIndex, 1);
      selectedTemplate = {...selectedTemplate}; // Trigger reactivity
      $templateEditorState = _.assign({}, $templateEditorState, { hasUnsavedChanges: true });
    }
  }
</script>

<Modal bind:active={templateCreateModalOpen}>
  <svelte:fragment slot="head" let:close>
    <p class="modal-card-title">Create Template</p>
    <button class="delete" aria-label="close" on:click={(e) => close(e)} />
  </svelte:fragment>
  <div class="field" slot="body">
    <label class="label" for="save-filename">Template Name</label>
    <div class="control" id="save-filename">
      <input class="input" type="text" bind:value={saveAsName} />
      {#if saveAsNameDuplicate}
        <p class="help is-danger">Template with the same name already exists</p>
      {/if}
    </div>
  </div>
  <svelte:fragment slot="foot" let:close>
    <button
      class="button is-success"
      disabled={_.isEmpty(saveAsName) || saveAsNameDuplicate}
      on:click={(e) => save() && close(e)}>Create</button
    >
    <button class="button" on:click={(e) => close(e)}>Cancel</button>
  </svelte:fragment>
</Modal>

<FileModal bind:open={modalOpen} on:save={(e) => saveToFile(e.detail)} />

<section class="section tab-import" style="padding-bottom: 0 !important">
  <div class="container is-fluid">
    <div class="columns mb-0">
      <div class="column is-5 py-0">
        <div class="box p-3 mb-3 overflow-x-auto">
          <div class="field is-grouped mb-0">
            <p class="control">
              <span data-tippy-content="Create" data-tippy-followCursor="false">
                <button class="button" on:click={(_e) => openTemplateCreateModal()}>
                  <span class="icon">
                    <i class="fas fa-file-circle-plus" />
                  </span>
                </button>
              </span>

              <span
                class="ml-4"
                data-tippy-followCursor="false"
                data-tippy-content={$templateEditorState.hasUnsavedChanges == false
                  ? "No Unsaved Chagnes"
                  : builtinNotAllowed("Save", selectedTemplate)}
              >
                <button
                  class="button"
                  on:click={(_e) => save()}
                  disabled={$templateEditorState.hasUnsavedChanges == false ||
                    selectedTemplate?.template_type == "builtin"}
                >
                  <span class="icon">
                    <i class="fas fa-floppy-disk" />
                  </span>
                </button>
              </span>

              <span
                data-tippy-followCursor="false"
                data-tippy-content={builtinNotAllowed("Delete", selectedTemplate)}
              >
                <button
                  class="button"
                  on:click={(_e) => remove()}
                  disabled={selectedTemplate?.template_type == "builtin"}
                >
                  <span class="icon">
                    <i class="fas fa-trash-can" />
                  </span>
                </button>
              </span>
            </p>

            <p class="control is-expanded">
              <Select
                bind:value={selectedTemplate}
                showChevron={true}
                items={templates}
                label="name"
                itemId="id"
                searchable={true}
                clearable={false}
                floatingConfig={{ strategy: "fixed" }}
                on:change={(_e) => {
                  saveAsName = selectedTemplate.name;
                }}
              >
                <div slot="selection" let:selection>
                  {selection.name}
                  <span class="tag is-small is-link invertable is-light"
                    >{selection.template_type}</span
                  >
                </div>
                <div slot="item" let:item>
                  <span class="name">{item.name}</span>
                  <span class="tag is-small is-link invertable is-light">{item.template_type}</span>
                </div>
              </Select>
            </p>
          </div>
        </div>
        <div class="box py-0">
          <div class="field">
            <div class="control">
              <div class="template-editor" bind:this={templateEditorDom} />
            </div>
          </div>
        </div>
        
        <!-- Rules Section -->
        <div class="box py-3">
          <div class="is-flex justify-between align-items-center mb-2">
            <h4 class="title is-5 mb-0">Transaction Rules</h4>
            <button 
              class="button is-small is-primary" 
              on:click={addRule}
              disabled={selectedTemplate?.template_type === "builtin"}>
              <span class="icon">
                <i class="fas fa-plus"></i>
              </span>
              <span>Add Rule</span>
            </button>
          </div>
          
          <p class="help mb-3">
            Rules allow you to filter transactions or add tags based on conditions. 
            Conditions are JavaScript expressions that have access to ROW, SHEET, and column references (A, B, C, etc.).
          </p>
          
          {#if selectedTemplate?.rules && selectedTemplate.rules.length > 0}
            {#each selectedTemplate.rules as rule, i}
              <div class="rule-container mb-4 p-3 has-background-light">
                <div class="is-flex justify-between align-items-center mb-2">
                  <h5 class="title is-6 mb-0">Rule #{i+1}</h5>
                  <button 
                    class="button is-small is-danger" 
                    on:click={() => removeRule(i)}
                    disabled={selectedTemplate?.template_type === "builtin"}>
                    <span class="icon">
                      <i class="fas fa-trash"></i>
                    </span>
                  </button>
                </div>
                
                <div class="field">
                  <label class="label">Condition</label>
                  <div class="control">
                    <input 
                      class="input" 
                      type="text" 
                      placeholder="e.g. ROW.B.includes('GROCERY') || ROW.C > 1000" 
                      bind:value={rule.condition}
                      on:input={(e) => updateRuleCondition(i, e.target.value)}
                      disabled={selectedTemplate?.template_type === "builtin"}
                    />
                  </div>
                  <p class="help">JavaScript expression that evaluates to true/false</p>
                </div>
                
                <div class="field">
                  <div class="control">
                    <label class="checkbox">
                      <input 
                        type="checkbox" 
                        bind:checked={rule.skip}
                        on:change={(e) => updateRuleSkip(i, e.target.checked)}
                        disabled={selectedTemplate?.template_type === "builtin"}
                      />
                      Skip transactions matching this condition
                    </label>
                  </div>
                </div>
                
                <div class="field">
                  <label class="label">Tags</label>
                  <div class="is-flex gap-2 flex-wrap mb-2">
                    {#if rule.tags && rule.tags.length > 0}
                      {#each rule.tags as tag, tagIndex}
                        <span class="tag is-info is-medium">
                          {tag}
                          <button 
                            class="delete is-small" 
                            on:click={() => removeTag(i, tagIndex)}
                            disabled={selectedTemplate?.template_type === "builtin"}
                          ></button>
                        </span>
                      {/each}
                    {:else}
                      <p class="help">No tags added yet</p>
                    {/if}
                  </div>
                  
                  <div class="field has-addons">
                    <div class="control is-expanded">
                      <input 
                        class="input" 
                        type="text" 
                        placeholder="Add a tag" 
                        bind:value={newTag}
                        disabled={selectedTemplate?.template_type === "builtin"}
                      />
                    </div>
                    <div class="control">
                      <button 
                        class="button is-info" 
                        on:click={() => {
                          addTag(i, newTag);
                          newTag = "";
                        }}
                        disabled={!newTag || selectedTemplate?.template_type === "builtin"}>
                        Add
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            {/each}
          {:else}
            <p class="has-text-centered py-4">No rules defined. Add a rule to filter transactions or add tags.</p>
          {/if}
        </div>
        <div class="box py-0">
          <div class="field">
            <div class="control">
              <button
                data-tippy-followCursor="false"
                data-tippy-content="Copy to Clipboard"
                class="button clipboard"
                disabled={_.isEmpty(preview)}
                on:click={copyToClipboard}
              >
                <span class="icon">
                  <i class="fas fa-copy" />
                </span>
              </button>
              <button
                data-tippy-followCursor="false"
                data-tippy-content="Save"
                class="button save"
                disabled={_.isEmpty(preview)}
                on:click={openSaveModal}
              >
                <span class="icon">
                  <i class="fas fa-floppy-disk" />
                </span>
              </button>
              <div class="preview-editor" bind:this={previewEditorDom} />
            </div>
          </div>
        </div>
      </div>
      <div class="column is-7 py-0">
        <div class="box p-3 mb-3">
          <Dropzone
            multiple={false}
            inputElement={input}
            accept=".csv,.txt,.xls,.xlsx,.pdf,.CSV,.TXT,.XLS,.XLSX,.PDF"
            on:drop={handleFilesSelect}
          >
            Drag 'n' drop CSV, TXT, XLS, XLSX, PDF file here or click to select
          </Dropzone>
        </div>
        <div class="is-flex justify-end mb-3 gap-4">
          <div class="field color-switch">
            <input
              id="import-reverse"
              type="checkbox"
              bind:checked={options.reverse}
              class="switch is-rounded is-small"
            />
            <label for="import-reverse">Reverse</label>
          </div>
          <div class="field color-switch">
            <input
              id="trim-reverse"
              type="checkbox"
              bind:checked={options.trim}
              class="switch is-rounded is-small"
            />
            <label for="trim-reverse">Trim</label>
          </div>
        </div>
        {#if parseErrorMessage}
          <div class="message invertable is-danger">
            <div class="message-header">Failed to parse document</div>
            <div class="message-body">{parseErrorMessage}</div>
          </div>
        {/if}
        {#if !_.isEmpty(data)}
          <div class="table-wrapper">
            <table
              class="mt-0 table is-bordered is-size-7 is-narrow has-sticky-header has-sticky-column"
            >
              <thead>
                <tr>
                  <th />
                  {#each _.range(0, columnCount) as ci}
                    <th class="has-background-light">{String.fromCharCode(65 + ci)}</th>
                  {/each}
                </tr>
              </thead>
              <tbody>
                {#each data as row, ri}
                  <tr>
                    <th class="has-background-light"><b>{ri}</b></th>
                    {#each row as cell}
                      <td>{cell || ""}</td>
                    {/each}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>
    </div>
    <div />
  </div>
</section>

<style lang="scss">
  @import "bulma/sass/utilities/_all.sass";

  $import-full-height: calc(100vh - 205px);

  .clipboard {
    float: right;
    position: absolute;
    right: 0;
    z-index: 1;
  }

  .save {
    float: right;
    position: absolute !important;
    right: 40px;
    z-index: 1;
  }

  .table-wrapper {
    overflow-x: auto;
    overflow-y: auto;
    max-height: $import-full-height;
  }

  .color-switch {
    .switch[type="checkbox"]:checked + label::before,
    .switch[type="checkbox"]:checked + label:before {
      background: $link;
    }
  }
  
  .rule-container {
    border-radius: 4px;
    border: 1px solid #dbdbdb;
  }
  
  .justify-between {
    justify-content: space-between;
  }
  
  .align-items-center {
    align-items: center;
  }
  
  .gap-2 {
    gap: 0.5rem;
  }
  
  .flex-wrap {
    flex-wrap: wrap;
  }
</style>
