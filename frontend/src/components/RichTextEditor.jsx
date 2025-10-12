import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Image from '@tiptap/extension-image'
import { useEffect } from 'react'

export default function RichTextEditor({ value, onChange, readOnly = false, placeholder = 'Start typing...' }) {
  const editor = useEditor({
    extensions: [
      StarterKit,
      Image,
    ],
    content: value || '',
    editable: !readOnly,
    onUpdate: ({ editor }) => {
      if (onChange) {
        onChange(editor.getJSON())
      }
    },
  })

  // Update editor content when value prop changes
  useEffect(() => {
    if (editor && value && JSON.stringify(editor.getJSON()) !== JSON.stringify(value)) {
      editor.commands.setContent(value)
    }
  }, [value, editor])

  if (!editor) {
    return null
  }

  if (readOnly) {
    return (
      <div className="prose max-w-none">
        <EditorContent editor={editor} />
      </div>
    )
  }

  return (
    <div className="border border-secondary-300 rounded-lg overflow-hidden">
      {/* Toolbar */}
      <div className="bg-secondary-50 border-b border-secondary-300 p-2 flex flex-wrap gap-1">
        <button
          type="button"
          onClick={() => editor.chain().focus().toggleBold().run()}
          className={`px-3 py-1 rounded text-sm font-medium ${
            editor.isActive('bold')
              ? 'bg-primary-100 text-primary-800'
              : 'bg-white text-secondary-700 hover:bg-secondary-100'
          }`}
        >
          <strong>B</strong>
        </button>
        <button
          type="button"
          onClick={() => editor.chain().focus().toggleItalic().run()}
          className={`px-3 py-1 rounded text-sm font-medium ${
            editor.isActive('italic')
              ? 'bg-primary-100 text-primary-800'
              : 'bg-white text-secondary-700 hover:bg-secondary-100'
          }`}
        >
          <em>I</em>
        </button>
        <button
          type="button"
          onClick={() => editor.chain().focus().toggleStrike().run()}
          className={`px-3 py-1 rounded text-sm font-medium ${
            editor.isActive('strike')
              ? 'bg-primary-100 text-primary-800'
              : 'bg-white text-secondary-700 hover:bg-secondary-100'
          }`}
        >
          <s>S</s>
        </button>
        
        <div className="w-px bg-secondary-300 mx-1"></div>
        
        <button
          type="button"
          onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()}
          className={`px-3 py-1 rounded text-sm font-medium ${
            editor.isActive('heading', { level: 1 })
              ? 'bg-primary-100 text-primary-800'
              : 'bg-white text-secondary-700 hover:bg-secondary-100'
          }`}
        >
          H1
        </button>
        <button
          type="button"
          onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
          className={`px-3 py-1 rounded text-sm font-medium ${
            editor.isActive('heading', { level: 2 })
              ? 'bg-primary-100 text-primary-800'
              : 'bg-white text-secondary-700 hover:bg-secondary-100'
          }`}
        >
          H2
        </button>
        <button
          type="button"
          onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}
          className={`px-3 py-1 rounded text-sm font-medium ${
            editor.isActive('heading', { level: 3 })
              ? 'bg-primary-100 text-primary-800'
              : 'bg-white text-secondary-700 hover:bg-secondary-100'
          }`}
        >
          H3
        </button>
        
        <div className="w-px bg-secondary-300 mx-1"></div>
        
        <button
          type="button"
          onClick={() => editor.chain().focus().toggleBulletList().run()}
          className={`px-3 py-1 rounded text-sm font-medium ${
            editor.isActive('bulletList')
              ? 'bg-primary-100 text-primary-800'
              : 'bg-white text-secondary-700 hover:bg-secondary-100'
          }`}
        >
          • List
        </button>
        <button
          type="button"
          onClick={() => editor.chain().focus().toggleOrderedList().run()}
          className={`px-3 py-1 rounded text-sm font-medium ${
            editor.isActive('orderedList')
              ? 'bg-primary-100 text-primary-800'
              : 'bg-white text-secondary-700 hover:bg-secondary-100'
          }`}
        >
          1. List
        </button>
        
        <div className="w-px bg-secondary-300 mx-1"></div>
        
        <button
          type="button"
          onClick={() => {
            const url = window.prompt('Enter image URL:')
            if (url) {
              editor.chain().focus().setImage({ src: url }).run()
            }
          }}
          className="px-3 py-1 rounded text-sm font-medium bg-white text-secondary-700 hover:bg-secondary-100"
        >
          🖼️ Image
        </button>
        
        <button
          type="button"
          onClick={() => editor.chain().focus().setHorizontalRule().run()}
          className="px-3 py-1 rounded text-sm font-medium bg-white text-secondary-700 hover:bg-secondary-100"
        >
          ─ HR
        </button>
        
        <div className="w-px bg-secondary-300 mx-1"></div>
        
        <button
          type="button"
          onClick={() => editor.chain().focus().undo().run()}
          disabled={!editor.can().undo()}
          className="px-3 py-1 rounded text-sm font-medium bg-white text-secondary-700 hover:bg-secondary-100 disabled:opacity-50"
        >
          ↶ Undo
        </button>
        <button
          type="button"
          onClick={() => editor.chain().focus().redo().run()}
          disabled={!editor.can().redo()}
          className="px-3 py-1 rounded text-sm font-medium bg-white text-secondary-700 hover:bg-secondary-100 disabled:opacity-50"
        >
          ↷ Redo
        </button>
      </div>

      {/* Editor Content */}
      <div className="p-4 min-h-[200px] prose max-w-none focus:outline-none">
        <EditorContent editor={editor} placeholder={placeholder} />
      </div>
    </div>
  )
}

