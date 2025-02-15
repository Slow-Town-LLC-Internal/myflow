complete -C /usr/local/bin/aws_completer aws


,get-gi () {
  aws ec2 describe-images \
    --profile $2 \
    --region ${1:-eu-central-2} \
    --output text --owners self \
    --filters "Name=name,Values=GoldenImage-Ubuntu-20*" \
    --query 'Images[*].[CreationDate,Name,ImageId] | [*] | sort_by(@, &[0])' \
    | tail -3 ;
}

,get_ec2id() {
  aws ec2 describe-instances --filters "Name=tag:Name,Values=$1"
}
,get_ec2info() {
  aws ec2 describe-instances --instance-id $1 --profile $2 --output ${3:-table}
}

,awslookup() {
  cmd="aws --profile $1 ec2 describe-instances --filters \"Name=tag:Name,Values=$2\" --query 'Reservations[].Instances[].[InstanceId,PublicDnsName,PrivateIpAddress,State.Name,Placement.AvailabilityZone,InstanceType,join(\`,\`,Tags[?Key==\`Name\`].Value)]' --output ${3:-table}"
  if [ $# -eq 4 ]
  then
    echo "Running $cmd"
  fi
  eval $cmd
}
,get_myinstances(){
    aws ec2  describe-instances \
      --filter "Name=tag:Role,Values=mysql"  \
      --query "Reservations[*].Instances[*].{name: Tags[?Key=='Name'] | [0].Value, id: InstanceId, type: InstanceType, zone:Placement.AvailabilityZone,IP: PrivateIpAddress, status: State.Name, mysql: Tags[?Key=='mysql-status'] |[0].Value } | sort_by(@, &[0].name)" \
      --output table --color off
}

,get_xtrasnap(){
    aws ec2 describe-snapshots \
      --filters Name=tag:created_by,Values="xtrabackup via ansible" \
      --output table \
      --query "Snapshots[*].{enc: Encrypted, id: SnapshotId, StartTime: StartTime,State:State,VolumeId:VolumeId,size:VolumeSize, name:Tags[?Key=='Name']|[0].Value,created:Tags[?Key=='created_by']|[0].Value}"
}
