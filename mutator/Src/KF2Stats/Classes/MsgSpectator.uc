class MsgSpectator extends MessagingSpectator;

reliable client event TeamMessage(PlayerReplicationInfo PRI, coerce string S, name Type, optional float MsgLifeTime) {
	ReceiveMessage(pri, s, type);
}

reliable client event ReceiveLocalizedMessage( class<LocalMessage> Message, optional int Switch, optional PlayerReplicationInfo RelatedPRI_1, optional PlayerReplicationInfo RelatedPRI_2, optional Object OptionalObject ) {
	if (ClassIsChildOf(Message, class'KFLocalMessage_VoiceComms')) {
		class'StatsRepo'.static.GetInstance().AddRadioComms(RelatedPRI_1, Switch);
	}
}