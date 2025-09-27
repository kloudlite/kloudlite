"use client";

import { useState } from "react";

import { CheckCircle, XCircle, Clock, MapPin } from "lucide-react";
import { useRouter } from "next/navigation";
import { type Session } from "next-auth";
import { toast } from "sonner";

import { approveTeamRequest, rejectTeamRequest } from "@/app/actions/teams";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";


interface TeamRequestsProps {
  requests: any[];
  session: Session;
}

export default function TeamRequests({ requests, session }: TeamRequestsProps) {
  const router = useRouter();
  const [selectedRequest, setSelectedRequest] = useState<any>(null);
  const [rejectReason, setRejectReason] = useState("");
  const [isProcessing, setIsProcessing] = useState(false);
  const [showRejectDialog, setShowRejectDialog] = useState(false);

  const handleApprove = async (requestId: string) => {
    setIsProcessing(true);
    try {
      await approveTeamRequest(requestId);
      toast.success("Team request approved successfully");
      router.refresh();
    } catch (error) {
      toast.error("Failed to approve team request");
    } finally {
      setIsProcessing(false);
    }
  };

  const handleReject = async () => {
    if (!selectedRequest || !rejectReason.trim()) {
      toast.error("Please provide a reason for rejection");
      return;
    }

    setIsProcessing(true);
    try {
      await rejectTeamRequest(selectedRequest.requestId, rejectReason);
      toast.success("Team request rejected");
      setShowRejectDialog(false);
      setSelectedRequest(null);
      setRejectReason("");
      router.refresh();
    } catch (error) {
      toast.error("Failed to reject team request");
    } finally {
      setIsProcessing(false);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "pending":
        return (
          <Badge variant="secondary" className="flex items-center gap-1">
            <Clock className="h-3 w-3" />
            Pending
          </Badge>
        );
      case "approved":
        return (
          <Badge variant="default" className="flex items-center gap-1">
            <CheckCircle className="h-3 w-3" />
            Approved
          </Badge>
        );
      case "rejected":
        return (
          <Badge variant="destructive" className="flex items-center gap-1">
            <XCircle className="h-3 w-3" />
            Rejected
          </Badge>
        );
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  if (requests.length === 0) {
    return (
      <Card className="glass-card">
        <CardHeader>
          <CardTitle>Team Requests</CardTitle>
          <CardDescription>
            Review and manage team creation requests
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">
            No pending team requests
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <Card className="glass-card">
        <CardHeader>
          <CardTitle>Team Requests</CardTitle>
          <CardDescription>
            Review and manage team creation requests
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[35%]">Team</TableHead>
                  <TableHead className="w-[25%]">Requested by</TableHead>
                  <TableHead className="w-[15%]">Status</TableHead>
                  <TableHead className="w-[25%] text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {requests.map((request) => (
                  <TableRow key={request.requestId}>
                    <TableCell className="py-4">
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          <p className="font-medium">{request.displayName}</p>
                          <span className="text-xs font-mono text-muted-foreground">
                            {request.slug}
                          </span>
                        </div>
                        {request.description && (
                          <p className="text-sm text-muted-foreground line-clamp-1">
                            {request.description}
                          </p>
                        )}
                        <div className="flex items-center gap-4 text-xs text-muted-foreground">
                          <div className="flex items-center gap-1">
                            <MapPin className="h-3 w-3" />
                            <span className="font-mono">{request.region}</span>
                          </div>
                          <span>•</span>
                          <span>{formatDate(request.requestedAt)}</span>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className="py-4">
                      <div className="text-sm">
                        <p>{request.requestedByEmail}</p>
                      </div>
                    </TableCell>
                    <TableCell className="py-4">
                      {getStatusBadge(request.status)}
                    </TableCell>
                    <TableCell className="py-4 text-right">
                      {request.status === "pending" && (
                        <div className="flex gap-2 justify-end">
                          <Button
                            size="sm"
                            onClick={() => handleApprove(request.requestId)}
                            disabled={isProcessing}
                          >
                            Approve
                          </Button>
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={() => {
                              setSelectedRequest(request);
                              setShowRejectDialog(true);
                            }}
                            disabled={isProcessing}
                          >
                            Reject
                          </Button>
                        </div>
                      )}
                      {request.status === "approved" && request.reviewedAt && (
                        <div className="text-xs text-muted-foreground text-right">
                          <p>Approved by {request.reviewedByEmail}</p>
                          <p>{formatDate(request.reviewedAt)}</p>
                        </div>
                      )}
                      {request.status === "rejected" && request.reviewedAt && (
                        <div className="text-xs text-muted-foreground text-right">
                          <p>Rejected by {request.reviewedByEmail}</p>
                          <p>{formatDate(request.reviewedAt)}</p>
                          {request.rejectionReason && (
                            <p className="mt-1 italic">"{request.rejectionReason}"</p>
                          )}
                        </div>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      <Dialog open={showRejectDialog} onOpenChange={setShowRejectDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reject Team Request</DialogTitle>
            <DialogDescription>
              Provide a reason for rejecting the team creation request for{" "}
              <strong>{selectedRequest?.displayName}</strong>.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label htmlFor="reason">Rejection Reason</Label>
              <Textarea
                id="reason"
                placeholder="Enter the reason for rejection..."
                value={rejectReason}
                onChange={(e) => setRejectReason(e.target.value)}
                className="mt-2"
                rows={4}
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setShowRejectDialog(false);
                setRejectReason("");
                setSelectedRequest(null);
              }}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleReject}
              disabled={isProcessing || !rejectReason.trim()}
            >
              Reject Request
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}