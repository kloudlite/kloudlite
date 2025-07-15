"use client"

import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { Upload, X, File, Image, FileText, Archive } from "lucide-react"

import { cn } from "@/lib/utils"
import { Button } from "./button"

const fileUploadVariants = cva(
  "relative flex flex-col items-center justify-center rounded-lg border-2 border-dashed border-input bg-background text-center transition-all duration-200 ease-in-out",
  {
    variants: {
      size: {
        sm: "h-32 p-4",
        default: "h-48 p-6",
        lg: "h-64 p-8",
      },
      state: {
        default: "hover:border-primary hover:bg-primary/5",
        dragover: "border-primary bg-primary/10 scale-[1.02]",
        error: "border-destructive bg-destructive/5",
        success: "border-success bg-success/5",
      },
    },
    defaultVariants: {
      size: "default",
      state: "default",
    },
  }
)

interface FileUploadProps
  extends Omit<React.HTMLAttributes<HTMLDivElement>, "onChange">,
    VariantProps<typeof fileUploadVariants> {
  onFileSelect?: (files: FileList | null) => void
  accept?: string
  multiple?: boolean
  maxSize?: number // in MB
  disabled?: boolean
  value?: FileList | File[] | null
  placeholder?: string
  showFileList?: boolean
}

function getFileIcon(file: File) {
  const type = file.type
  if (type.startsWith("image/")) return Image
  if (type.includes("pdf") || type.includes("text")) return FileText
  if (type.includes("zip") || type.includes("rar")) return Archive
  return File
}

function formatFileSize(bytes: number) {
  if (bytes === 0) return "0 Bytes"
  const k = 1024
  const sizes = ["Bytes", "KB", "MB", "GB"]
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i]
}

const FileUpload = React.forwardRef<HTMLDivElement, FileUploadProps>(
  ({
    className,
    size,
    state: stateProp,
    onFileSelect,
    accept,
    multiple = false,
    maxSize,
    disabled = false,
    value,
    placeholder = "Drop files here or click to browse",
    showFileList = true,
    ...props
  }, ref) => {
    const [dragState, setDragState] = React.useState<"default" | "dragover">("default")
    const [error, setError] = React.useState<string | null>(null)
    const [files, setFiles] = React.useState<File[]>([])
    const inputRef = React.useRef<HTMLInputElement>(null)
    const dragCounter = React.useRef(0)

    const currentState = stateProp || (error ? "error" : files.length > 0 ? "success" : dragState)

    React.useEffect(() => {
      if (value) {
        const fileArray = value instanceof FileList ? Array.from(value) : Array.isArray(value) ? value : [value]
        setFiles(fileArray)
        setError(null)
      } else {
        // Reset files state when value is null/empty
        setFiles([])
        setError(null)
      }
    }, [value])

    const validateFiles = (fileList: FileList): { validFiles: File[], errors: string[] } => {
      const validFiles: File[] = []
      const errors: string[] = []

      Array.from(fileList).forEach(file => {
        if (maxSize && file.size > maxSize * 1024 * 1024) {
          errors.push(`${file.name}: File size exceeds ${maxSize}MB`)
          return
        }
        
        if (accept) {
          const acceptedTypes = accept.split(',').map(type => type.trim())
          const isAccepted = acceptedTypes.some(type => {
            if (type.startsWith('.')) {
              return file.name.toLowerCase().endsWith(type.toLowerCase())
            }
            return file.type.match(type.replace('*', '.*'))
          })
          
          if (!isAccepted) {
            errors.push(`${file.name}: File type not accepted`)
            return
          }
        }
        
        validFiles.push(file)
      })

      return { validFiles, errors }
    }

    const handleFiles = (fileList: FileList | null) => {
      if (!fileList || fileList.length === 0) return

      const { validFiles, errors } = validateFiles(fileList)
      
      if (errors.length > 0) {
        setError(errors.join(', '))
        return
      }

      setError(null)
      setFiles(multiple ? [...files, ...validFiles] : validFiles)
      onFileSelect?.(multiple ? 
        (() => {
          const dt = new DataTransfer()
          ;[...files, ...validFiles].forEach(file => dt.items.add(file))
          return dt.files
        })() : 
        (() => {
          const dt = new DataTransfer()
          validFiles.forEach(file => dt.items.add(file))
          return dt.files
        })()
      )
    }

    const handleDragEnter = (e: React.DragEvent) => {
      e.preventDefault()
      e.stopPropagation()
      dragCounter.current++
      if (e.dataTransfer.items && e.dataTransfer.items.length > 0) {
        setDragState("dragover")
      }
    }

    const handleDragLeave = (e: React.DragEvent) => {
      e.preventDefault()
      e.stopPropagation()
      dragCounter.current--
      if (dragCounter.current === 0) {
        setDragState("default")
      }
    }

    const handleDragOver = (e: React.DragEvent) => {
      e.preventDefault()
      e.stopPropagation()
    }

    const handleDrop = (e: React.DragEvent) => {
      e.preventDefault()
      e.stopPropagation()
      setDragState("default")
      dragCounter.current = 0
      
      if (disabled) return
      
      const droppedFiles = e.dataTransfer.files
      handleFiles(droppedFiles)
    }

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      handleFiles(e.target.files)
      // Clear the input value to allow selecting the same file again
      e.target.value = ''
    }

    const handleClick = () => {
      if (!disabled) {
        inputRef.current?.click()
      }
    }

    const removeFile = (index: number) => {
      const newFiles = files.filter((_, i) => i !== index)
      setFiles(newFiles)
      
      if (newFiles.length === 0) {
        onFileSelect?.(null)
        setError(null)
      } else {
        const dt = new DataTransfer()
        newFiles.forEach(file => dt.items.add(file))
        onFileSelect?.(dt.files)
      }
    }

    return (
      <div className="space-y-4">
        <div
          ref={ref}
          className={cn(
            fileUploadVariants({ size, state: currentState }),
            disabled && "opacity-50 cursor-not-allowed",
            !disabled && "cursor-pointer",
            className
          )}
          onDragEnter={handleDragEnter}
          onDragLeave={handleDragLeave}
          onDragOver={handleDragOver}
          onDrop={handleDrop}
          onClick={handleClick}
          {...props}
        >
          <input
            ref={inputRef}
            type="file"
            className="hidden"
            accept={accept}
            multiple={multiple}
            onChange={handleInputChange}
            disabled={disabled}
          />
          
          <div className="flex flex-col items-center gap-2">
            <Upload className={cn(
              "transition-all duration-200",
              size === "sm" ? "h-6 w-6" : size === "lg" ? "h-10 w-10" : "h-8 w-8",
              currentState === "dragover" ? "text-primary scale-110" : "text-muted-foreground",
              currentState === "error" ? "text-destructive" : "",
              currentState === "success" ? "text-success" : ""
            )} />
            
            <div className="space-y-1">
              <p className={cn(
                "text-sm font-medium transition-colors duration-200",
                currentState === "dragover" ? "text-primary" : "text-foreground",
                currentState === "error" ? "text-destructive" : "",
                currentState === "success" ? "text-success" : ""
              )}>
                {currentState === "dragover" ? "Drop files here" : placeholder}
              </p>
              
              {!disabled && (
                <p className="text-xs text-muted-foreground">
                  {accept && `Accepted: ${accept}`}
                  {maxSize && ` • Max: ${maxSize}MB`}
                  {multiple && " • Multiple files allowed"}
                </p>
              )}
            </div>
          </div>
        </div>

        {error && (
          <div className="text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-md p-3">
            {error}
          </div>
        )}

        {showFileList && files.length > 0 && (
          <div className="space-y-2">
            <h4 className="text-sm font-medium text-foreground">
              Selected files ({files.length})
            </h4>
            <div className="space-y-2">
              {files.map((file, index) => {
                const FileIcon = getFileIcon(file)
                return (
                  <div
                    key={`${file.name}-${index}`}
                    className="flex items-center gap-3 p-3 bg-muted/50 rounded-lg border border-border"
                  >
                    <FileIcon className="h-4 w-4 text-muted-foreground shrink-0" />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-foreground truncate">
                        {file.name}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {formatFileSize(file.size)}
                      </p>
                    </div>
                    {!disabled && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation()
                          removeFile(index)
                        }}
                        className="h-auto p-1 hover:bg-destructive/10 hover:text-destructive"
                      >
                        <X className="h-3 w-3" />
                      </Button>
                    )}
                  </div>
                )
              })}
            </div>
          </div>
        )}
      </div>
    )
  }
)

FileUpload.displayName = "FileUpload"

export { FileUpload, fileUploadVariants }
export type { FileUploadProps }