'use client'

import { useEffect, useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { File, Upload, Trash2, Download, AlertCircle, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { listFiles, getFile, deleteFile } from '@/app/actions/environment-config'
import type { FileInfo } from '@/types/environment'
import { AddFileSheet } from '@/components/add-file-sheet'
import { EditFileSheet } from '@/components/edit-file-sheet'
import { toast } from 'sonner'

interface FilesListProps {
  environmentId: string
}

export function FilesList({ environmentId }: FilesListProps) {
  const router = useRouter()
  const [files, setFiles] = useState<FileInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  // Delete dialog state
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [selectedFile, setSelectedFile] = useState<FileInfo | null>(null)

  useEffect(() => {
    loadFilesList()
  }, [environmentId])

  const loadFilesList = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await listFiles(environmentId)
      setFiles(response.files || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load files')
      setFiles([])
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteClick = (file: FileInfo) => {
    setSelectedFile(file)
    setShowDeleteDialog(true)
  }

  const handleDownload = async (file: FileInfo) => {
    try {
      const response = await getFile(environmentId, file.name)
      const blob = new Blob([response.content], { type: 'text/plain' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = file.name
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to download file')
    }
  }

  const handleDelete = () => {
    if (!selectedFile) return

    startTransition(async () => {
      try {
        await deleteFile(environmentId, selectedFile.name)
        toast.success('File deleted successfully')
        setShowDeleteDialog(false)
        router.refresh()
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to delete file')
      }
    })
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 border border-red-200 p-4">
        <div className="flex items-center gap-2 text-red-800">
          <AlertCircle className="h-5 w-5" />
          <span className="font-medium">Error loading files</span>
        </div>
        <p className="mt-2 text-sm text-red-700">{error}</p>
        <Button onClick={loadFilesList} variant="outline" size="sm" className="mt-3">
          Retry
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-medium">File Configs</h3>
          <p className="text-sm text-gray-500">Configuration files mounted to containers</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm">
            <Upload className="h-4 w-4 mr-2" />
            Upload File
          </Button>
          <AddFileSheet environmentId={environmentId} />
        </div>
      </div>

      {files.length === 0 ? (
        <div className="text-center py-12 bg-gray-50 rounded-lg border border-gray-200">
          <File className="h-12 w-12 mx-auto text-gray-400 mb-4" />
          <p className="text-gray-500">No configuration files</p>
          <div className="mt-4">
            <AddFileSheet environmentId={environmentId} />
          </div>
        </div>
      ) : (
        <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
          <table className="min-w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">File Name</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {files.map((file) => (
                <tr key={file.name} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    <div className="flex items-center gap-2">
                      <File className="h-4 w-4 text-gray-400" />
                      {file.name}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm space-x-1">
                    <Button variant="ghost" size="sm" onClick={() => handleDownload(file)}>
                      <Download className="h-4 w-4" />
                    </Button>
                    <EditFileSheet
                      environmentId={environmentId}
                      filename={file.name}
                    />
                    <Button variant="ghost" size="sm" className="text-red-600 hover:text-red-700" onClick={() => handleDeleteClick(file)}>
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Delete Dialog */}
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete File</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the file <span className="font-mono font-semibold">{selectedFile?.name}</span>?
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isPending}>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} disabled={isPending} className="bg-red-600 hover:bg-red-700">
              {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
